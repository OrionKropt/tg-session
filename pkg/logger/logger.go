package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
)

const (
	INFO = iota
	DEBUG
	WARN
	ERROR
)

type LogType int

type LogHandlerOptions struct {
	SlogOpts slog.HandlerOptions
}

type LogHandler struct {
	slog.Handler
	logger *log.Logger
}

type Logger struct {
	Inst *slog.Logger
	Name string
}

func (h *LogHandler) Handle(ctx context.Context, record slog.Record) error {
	level := record.Level.String() + ":"

	fields := make(map[string]interface{}, record.NumAttrs())
	record.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()

		return true
	})

	bytes, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return err
	}

	timeStr := record.Time.Format("02-01-2006 15:04:05")

	h.logger.Println(timeStr, level, record.Message, string(bytes))
	return nil
}

func NewLogHandler(out io.Writer, opts LogHandlerOptions) *LogHandler {
	handler := &LogHandler{
		Handler: slog.NewJSONHandler(out, &opts.SlogOpts),
		logger:  log.New(out, "", 0),
	}

	return handler
}

func ParseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func (l *Logger) Log(t LogType, message string, args ...any) {
	logger := l.Inst
	message = fmt.Sprintf("[%s] %s", l.Name, message)
	switch t {
	case INFO:
		logger.Info(message, args...)
	case DEBUG:
		logger.Debug(message, args...)
	case WARN:
		logger.Warn(message, args...)
	case ERROR:
		logger.Error(message, args...)
	default:
		logger.Debug(message, args...)
	}
}
