package mqdto

import "encoding/json"

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

func UnmarshallDeleteSubject(data []byte) (*DeleteSubject, error) {
	var ds DeleteSubject
	if err := json.Unmarshal(data, &ds); err != nil {
		return nil, err
	}
	return &ds, nil
}
