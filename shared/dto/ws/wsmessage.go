package wsdto

import "encoding/json"

type WSMessage struct {
	Type Operation       `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (wsm *WSMessage) GetBytes() ([]byte, error) {
	return json.Marshal(wsm)
}
