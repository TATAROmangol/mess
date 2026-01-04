package domain

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type AvatarIdentifier struct {
	SubjectID   string  `json:"subject_id"`
	ID          string  `json:"id"`
	PreviousKey *string `json:"previous_key,omitempty"`
}

func NewAvatarIdentifier(subjectID string, prevToken *string) *AvatarIdentifier {
	id := uuid.New().String()

	return &AvatarIdentifier{
		SubjectID:   subjectID,
		ID:          id,
		PreviousKey: prevToken,
	}
}

func (av *AvatarIdentifier) Key() (string, error) {
	data, err := json.Marshal(av)
	if err != nil {
		return "", fmt.Errorf("marshal: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func ParseAvatarKey(src string) (*AvatarIdentifier, error) {
	raw, err := base64.RawURLEncoding.DecodeString(src)
	if err != nil {
		return nil, errors.New("invalid token: cannot decode base64")
	}

	var av AvatarIdentifier
	if err := json.Unmarshal(raw, &av); err != nil {
		return nil, errors.New("invalid token: cannot unmarshal json")
	}

	if av.SubjectID == "" || av.ID == "" {
		return nil, errors.New("invalid token: missing required fields")
	}

	return &av, nil
}
