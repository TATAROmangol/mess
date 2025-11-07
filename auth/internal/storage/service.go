package storage

import (
	"auth/internal/entities"
	"context"
)

type Service interface {
	CreateSubject(ctx context.Context, login, password string) (int, error)
	GetSubjectByID(ctx context.Context, subjID int) (*entities.Subject, error)
	DeleteSubjectByID(ctx context.Context, subjID int) error
	ChangePassword(ctx context.Context, subjID int, newPassword string) error
}
