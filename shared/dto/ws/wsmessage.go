package ws

import "encoding/json"

type WSMessage struct {
	Type Operation       `json:"type"`
	Data json.RawMessage `json:"data"`
}
