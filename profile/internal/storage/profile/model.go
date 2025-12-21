package profile

import (
	"profile/internal/model"
	"time"
)

type ProfileEntity struct {
	SubjectID string    `db:"subject_id"`
	Alias     string    `db:"alias"`
	AvatarURL string    `db:"avatar_url"`
	Version   int       `db:"version"`
	UpdatedAt time.Time `db:"updated_at"`
	CreatedAt time.Time `db:"created_at"`
}

func (p *ProfileEntity) ToModel() *model.Profile {
	return &model.Profile{
		SubjectID: p.SubjectID,
		Alias:     p.Alias,
		AvatarURL: p.AvatarURL,
		Version:   p.Version,
		UpdatedAt: p.UpdatedAt,
		CreatedAt: p.CreatedAt,
	}
}

func (p *ProfileEntity) Key() *string {
	return &p.SubjectID
}

func ProfileEntitiesToModels(entities []*ProfileEntity) []*model.Profile {
	models := make([]*model.Profile, 0, len(entities))
	for _, entity := range entities {
		models = append(models, entity.ToModel())
	}
	return models
}
