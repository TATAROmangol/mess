package model

import wsdto "github.com/TATAROmangol/mess/shared/dto/ws"

type Message struct {
	SubjectID string
	WSMessage *wsdto.WSMessage
}
