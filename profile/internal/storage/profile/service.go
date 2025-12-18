package profile

import "profile/internal/model"

type Service interface {
	GetProfileFromSubjectID(subjID string) (*model.Profile, error)
}
