package profile

import "profile/internal/model"

type Service interface {
	AddProfile(profile *model.Profile) error
	GetProfileFromSubjectID(subjID string) (*model.Profile, error)
	UpdateProfile(profile *model.Profile) error
}
