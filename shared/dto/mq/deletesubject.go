package mqdto

type DeleteSubject struct {
	UserID     string `json:"userId"`
	ResourceID string `json:"resourceId"`
}

func (ds *DeleteSubject) GetSubjectID() string {
	if ds.ResourceID != "" {
		return ds.ResourceID
	}
	return ds.UserID
}
