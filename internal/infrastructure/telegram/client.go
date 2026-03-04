package telegram

import (
	"context"
	"fmt"

	pebbledb "github.com/cockroachdb/pebble"
	"github.com/gotd/contrib/bg"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/pebble"
	"github.com/gotd/contrib/storage"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"

	"log/slog"
	"tg-session/pkg/logger"

	"tg-session/internal/domain"
	"time"
)

type TGClient struct {
	log      logger.Logger
	client   *telegram.Client
	waiter   *floodwait.Waiter
	sender   *message.Sender
	stopFunc func() error
	updates  chan domain.Message
}

func (t *TGClient) Start() (err error) {
	errChan := make(chan error, 1)
	go func() {
		if err := t.waiter.Run(context.Background(), func(ctx context.Context) error {
			stop, err := bg.Connect(t.client)
			if err != nil {
				return err
			}
			t.stopFunc = stop
			errChan <- nil

			<-ctx.Done()
			return nil
		}); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-time.After(20 * time.Second):
		return fmt.Errorf("connection timeout")
	}
}

func (t *TGClient) Stop(ctx context.Context) error {
	status, err := t.client.Auth().Status(ctx)
	if err != nil {
		return err
	}
	if status.Authorized {
		if _, err := t.client.API().AuthLogOut(ctx); err != nil {
			return err
		}
		t.log.Log(logger.INFO, "logged out successfully")
	}
	close(t.updates)
	return t.stopFunc()
}

func findMessageID(arr []tg.UpdateClass) (int64, bool) {
	for _, update := range arr {
		if uID, ok := update.(*tg.UpdateMessageID); ok {
			return int64(uID.ID), true
		}
		if msg, ok := update.(*tg.UpdateNewMessage); ok {
			if m, ok := msg.Message.(*tg.Message); ok {
				return int64(m.ID), true
			}
		}
	}
	return 0, false
}

func extractMsgID(upd tg.UpdatesClass) (int64, bool) {
	switch u := upd.(type) {
	case *tg.UpdateShortSentMessage:
		return int64(u.ID), true
	case *tg.UpdateShortMessage:
		return int64(u.ID), true
	case *tg.UpdateShortChatMessage:
		return int64(u.ID), true
	case *tg.Updates:
		return findMessageID(u.Updates)
	case *tg.UpdatesCombined:
		return findMessageID(u.Updates)
	}

	return 0, false
}

func (t *TGClient) SendText(ctx context.Context, username, text string) (int64, error) {
	res, err := t.sender.ResolveDomain(username).Text(ctx, text)
	if err != nil {
		t.log.Log(logger.ERROR, "failed to send text", "username", username, "text", text)
		return 0, fmt.Errorf("failed to send text %w", err)
	}
	id, ok := extractMsgID(res)
	if !ok {
		err := fmt.Errorf("failed to resolve message ID")
		t.log.Log(logger.ERROR, err.Error(), "username", username, "text", text)
		return 0, err
	}
	return id, nil
}

func (t *TGClient) Updates() <-chan domain.Message {
	return t.updates
}

func (t *TGClient) authQRCode(ctx context.Context, tokenChan chan string) {
	go func() {
		authCtx, authCancel := context.WithTimeout(ctx, 5*time.Minute)
		defer authCancel()

		qr := t.client.QR()
		auth, err := qr.Auth(authCtx, make(chan struct{}), func(ctx context.Context, token qrlogin.Token) error {
			select {
			case tokenChan <- token.URL():
			default:

			}
			return nil
		})

		if err != nil {
			t.log.Log(logger.ERROR, "QR login failed ", "err", err.Error())
			return
		}
		if user, ok := auth.GetUser().AsNotEmpty(); ok {
			t.log.Log(logger.INFO, "authorized via QR finished", "username", user.Username)
		}
	}()
}

func (t *TGClient) AuthQRCode(ctx context.Context) (string, error) {
	tokenChan := make(chan string, 1)
	t.authQRCode(ctx, tokenChan)
	select {
	case url := <-tokenChan:
		return url, nil
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(time.Second * 30):
		return "", fmt.Errorf("timeout waiting for telegram qr")
	}
}

type TGClientFactory struct {
	appID   int
	appHash string
}

func NewTGClientFactory(id int, hash string) TGClientFactory {
	return TGClientFactory{
		appID:   id,
		appHash: hash,
	}
}

func newPeerDB(pebbleDBName string) (*pebble.PeerStorage, error) {
	db, err := pebbledb.Open(pebbleDBName, &pebbledb.Options{})
	if err != nil {
		return nil, err
	}
	return pebble.NewPeerStorage(db), nil
}

func newUpdateDispatcher(peerDB *pebble.PeerStorage, updatesChan chan domain.Message) (telegram.UpdateHandler, error) {
	dispatcher := tg.NewUpdateDispatcher()
	hook := storage.UpdateHook(dispatcher, peerDB)

	gaps := updates.New(updates.Config{
		Handler: hook,
	})

	dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
		msg, ok := u.Message.(*tg.Message)
		if !ok || msg.Out {
			return nil
		}
		p, err := storage.FindPeer(ctx, peerDB, msg.GetPeerID())
		senderName := "unknown"
		if err != nil {
			return err
		}

		if p.User != nil {
			senderName = fmt.Sprintf("@%s", p.User.Username)
		} else if p.Channel != nil {
			senderName = fmt.Sprintf("@%s", p.Channel.Username)
		}

		updatesChan <- domain.NewMessage(domain.MessageID(msg.ID), senderName, msg.Message)

		return nil
	})
	return gaps, nil
}

func (f *TGClientFactory) CreateTGClient(peerStorageDBName string, log *slog.Logger) (domain.TGClient, error) {
	waiter := floodwait.NewWaiter()
	updatesChan := make(chan domain.Message, 256)
	peerDB, err := newPeerDB(peerStorageDBName)
	if err != nil {
		return nil, err
	}
	dispatcher, err := newUpdateDispatcher(peerDB, updatesChan)
	if err != nil {
		return &TGClient{}, err
	}
	opts := telegram.Options{
		UpdateHandler:  dispatcher,
		SessionStorage: &session.StorageMemory{},
		DialTimeout:    5 * time.Second,
		MaxRetries:     3,
		RetryInterval:  2 * time.Second,
		Middlewares: []telegram.Middleware{
			waiter,
		}}

	client := telegram.NewClient(f.appID, f.appHash, opts)
	resolver := storage.NewResolverCache(peer.Plain(client.API()), peerDB)
	sender := message.NewSender(client.API()).WithResolver(resolver)

	return &TGClient{log: logger.Logger{Inst: log, Name: "TG Client"},
			updates: updatesChan,
			waiter:  waiter,
			client:  client,
			sender:  sender},
		nil
}
