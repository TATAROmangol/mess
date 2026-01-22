package wsdto

import (
	"encoding/json"
	"time"
)

type Message struct {
	ChatID    int       `json:"chat_id"`
	SenderID  string    `json:"sender_id"`
	Content   string    `json:"content"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
}

func (m *Message) GetData() ([]byte, error) {
	return json.Marshal(m)
}
