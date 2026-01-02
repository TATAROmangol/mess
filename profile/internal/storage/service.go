package storage

import (
	"context"
	"fmt"

	"github.com/TATAROmangol/mess/profile/internal/model"
	"github.com/TATAROmangol/mess/shared/postgres"
	"github.com/jmoiron/sqlx"
)

type Profile interface {
	AddProfile(ctx context.Context, subjID string, alias string) (*model.Profile, error)

	GetProfileFromSubjectID(ctx context.Context, subjID string) (*model.Profile, error)
	GetProfilesFromAlias(ctx context.Context, size int, asc bool, sortLabel Label, alias string) (string, []*model.Profile, error)
	GetProfilesFromAliasWithToken(ctx context.Context, token, alias string) (string, []*model.Profile, error)

	UpdateProfileMetadata(ctx context.Context, subjectID string, prevVersion int, alias string) (*model.Profile, error)
	UpdateAvatarKey(ctx context.Context, subjID string, avatarKey string) error

	DeleteProfile(ctx context.Context, subjID string) error
	DeleteAvatarKey(ctx context.Context, subjID string) error
}

type AvatarKeyOutbox interface {
	GetKeys(ctx context.Context, limit int) ([]*model.AvatarKeyOutbox, error)
	AddKey(ctx context.Context, subjectID string, key string) (*model.AvatarKeyOutbox, error)
	DeleteKeys(ctx context.Context, keys []string) error
}

type Service interface {
	WithTransaction(ctx context.Context) (ServiceTransaction, error)
	Profile
	AvatarKeyOutbox
}

type ServiceTransaction interface {
	Profile
	AvatarKeyOutbox
	Commit() error
	Rollback() error
}

type Storage struct {
	db   *sqlx.DB
	exec sqlx.ExtContext
}

func New(cfg postgres.Config) (*Storage, error) {
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
		return nil, fmt.Errorf("begin txx: %v", err)
	}

	return &Storage{
		db:   s.db,
		exec: tx,
	}, nil
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
