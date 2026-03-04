package domain

import "time"

type MessageID int64

type Message struct {
	ID        MessageID
	From      string
	Text      string
	CreatedAt time.Time
}

func NewMessage(id MessageID, from, text string) Message {
	return Message{ID: id, From: from, Text: text, CreatedAt: time.Now()}
}
