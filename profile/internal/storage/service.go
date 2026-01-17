package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/TATAROmangol/mess/profile/internal/model"
	"github.com/TATAROmangol/mess/shared/postgres"
	"github.com/jmoiron/sqlx"
)

type ProfilePaginationFilter struct {
	LastID    *string
	Limit     int
	Asc       bool
	SortLabel string
}

type Profile interface {
	AddProfile(ctx context.Context, subjID string, alias string) (*model.Profile, error)

	GetProfileFromSubjectID(ctx context.Context, subjID string) (*model.Profile, error)
	GetProfilesFromAlias(ctx context.Context, alias string, filter *ProfilePaginationFilter) ([]*model.Profile, error)

	UpdateProfileMetadata(ctx context.Context, subjectID string, prevVersion int, alias string) (*model.Profile, error)
	UpdateAvatarKey(ctx context.Context, subjID string, avatarKey string) (*model.Profile, error)

	DeleteProfile(ctx context.Context, subjID string) (*model.Profile, error)
	DeleteAvatarKey(ctx context.Context, subjID string) (*model.Profile, error)
}

type AvatarOutbox interface {
	GetKeys(ctx context.Context, limit int) ([]*model.AvatarOutbox, error)
	AddKey(ctx context.Context, subjectID string, key string) (*model.AvatarOutbox, error)
	DeleteKeys(ctx context.Context, keys []string) ([]*model.AvatarOutbox, error)
}

type Service interface {
	WithTransaction(ctx context.Context) (ServiceTransaction, error)
	Profile() Profile
	AvatarOutbox() AvatarOutbox
}

type ServiceTransaction interface {
	Profile() Profile
	AvatarOutbox() AvatarOutbox
	Commit() error
	Rollback() error
}

var (
	ErrNoRows = sql.ErrNoRows
)

type Storage struct {
	db   *sqlx.DB
	exec sqlx.ExtContext
}

func New(cfg postgres.Config) (Service, error) {
	db, err := postgres.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	return &Storage{
		db:   db,
		exec: db,
	}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) WithTransaction(ctx context.Context) (ServiceTransaction, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin txx: %w", err)
	}

	return &Storage{
		db:   s.db,
		exec: tx,
	}, nil
}

func (s *Storage) Profile() Profile {
	return &Storage{
		db:   s.db,
		exec: s.exec,
	}
}

func (s *Storage) AvatarOutbox() AvatarOutbox {
	return &Storage{
		db:   s.db,
		exec: s.exec,
	}
}

func (s *Storage) Commit() error {
	tx, ok := s.exec.(*sqlx.Tx)
	if !ok {
		return fmt.Errorf("commit called without transaction")
	}
	return tx.Commit()
}

func (s *Storage) Rollback() error {
	tx, ok := s.exec.(*sqlx.Tx)
	if !ok {
		return fmt.Errorf("rollback called without transaction")
	}
	return tx.Rollback()
}
