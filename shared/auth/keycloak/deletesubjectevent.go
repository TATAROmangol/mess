package keycloak

type ClientSubjectDeleteMessage struct {
	SubjectID string `json:"userId"`
}

func (cpdm *ClientSubjectDeleteMessage) GetSubjectID() string {
	return cpdm.SubjectID
}

type AdminSubjectDeleteMessage struct {
	SubjectID string `json:"resourceId"`
}

func (apdm *AdminSubjectDeleteMessage) GetSubjectID() string {
	return apdm.SubjectID
}
