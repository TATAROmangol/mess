package storage

import (
	"time"

	"github.com/TATAROmangol/mess/profile/internal/model"
)

type ProfileEntity struct {
	SubjectID string     `db:"subject_id"`
	Alias     string     `db:"alias"`
	AvatarKey *string     `db:"avatar_key"`
	Version   int        `db:"version"`
	UpdatedAt time.Time  `db:"updated_at"`
	CreatedAt time.Time  `db:"created_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func (p *ProfileEntity) ToModel() *model.Profile {
	return &model.Profile{
		SubjectID: p.SubjectID,
		Alias:     p.Alias,
		AvatarKey: p.AvatarKey,
		Version:   p.Version,
		UpdatedAt: p.UpdatedAt,
		CreatedAt: p.CreatedAt,
		DeletedAt: p.DeletedAt,
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

type AvatarKeyOutboxEntity struct {
	SubjectID string     `db:"subject_id"`
	KeyLabel  string     `db:"key"`
	CreatedAt time.Time  `db:"created_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func (p *AvatarKeyOutboxEntity) ToModel() *model.AvatarKeyOutbox {
	return &model.AvatarKeyOutbox{
		SubjectID: p.SubjectID,
		Key:       p.KeyLabel,
		DeletedAt: p.DeletedAt,
		CreatedAt: p.CreatedAt,
	}
}

func (p *AvatarKeyOutboxEntity) Key() *string {
	return &p.KeyLabel
}

func AvatarKeyOutboxEntitiesToModels(entities []*AvatarKeyOutboxEntity) []*model.AvatarKeyOutbox {
	models := make([]*model.AvatarKeyOutbox, 0, len(entities))
	for _, entity := range entities {
		models = append(models, entity.ToModel())
	}
	return models
}
