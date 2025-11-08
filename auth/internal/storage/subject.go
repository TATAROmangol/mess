package storage

import (
	"auth/internal/entities"
	"context"
	"fmt"
	"time"

	sq "github.com/doug-martin/goqu/v9"
)

type SubjectModel struct {
	ID           int        `db:"id"`
	Login        string     `db:"login"`
	PasswordHash string     `db:"password_hash"`
	CreatedAt    time.Time  `db:"created_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
	Version      int        `db:"version"`
}

func (sm *SubjectModel) ToEntity() *entities.Subject {
	return &entities.Subject{
		ID:           sm.ID,
		Login:        sm.Login,
		PasswordHash: sm.PasswordHash,
		CreatedAt:    sm.CreatedAt,
		DeletedAt:    sm.DeletedAt,
		Version:      sm.Version,
	}
}

var (
	ErrorNotFound = fmt.Errorf("subject not found")
	ErrorLastHash = fmt.Errorf("new password matches the old one")
)

func (s *Storage) GetSubjectByID(ctx context.Context, subjID int) (*entities.Subject, error) {
	query := sq.From(
		SubjectTableField,
	).Where(sq.Ex{
		IDSubjectColumnField: subjID,
		DeletedAtColumnField: nil,
	})

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("to sql: %w", err)
	}

	var sm SubjectModel
	err = s.db.GetContext(ctx, &sm, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("query row context: %w", err)
	}

	return sm.ToEntity(), nil
}

func (s *Storage) CreateSubject(ctx context.Context, login, password string) (int, error) {
	query := sq.Insert(
		SubjectTableField,
	).Rows(sq.Record{
		LoginColumnField:        login,
		PasswordHashColumnField: password,
		CreatedAtColumnField:    time.Now(),
	}).Returning(
		IDSubjectColumnField,
	)

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		return 0, fmt.Errorf("to sql: %w", err)
	}

	var id int
	err = s.db.QueryRowContext(ctx, sqlStr, args...).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("query row context: %w", err)
	}

	return id, nil
}

func (s *Storage) DeleteSubject(ctx context.Context, subjID int) error {
	query := sq.Update(
		SubjectTableField,
	).Set(
		sq.Record{
			DeletedAtColumnField: time.Now(),
		},
	).Where(
		sq.Ex{
			IDSubjectColumnField: subjID,
		},
	)

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		return fmt.Errorf("to sql: %w", err)
	}

	res, err := s.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if affected == 0 {
		return ErrorNotFound
	}

	return nil
}

func (s *Storage) ChangePasswordHash(ctx context.Context, subjID int, newPassword string) error {
	old, err := s.GetSubjectByID(ctx, subjID)
	if err != nil {
		return fmt.Errorf("get subject by id: %w", err)
	}

	if old.PasswordHash == newPassword {
		return ErrorLastHash
	}

	query := sq.Update(
		SubjectTableField,
	).Set(
		sq.Record{
			PasswordHashColumnField: newPassword,
			VersionField:            old.Version + 1,
		},
	).Where(
		sq.Ex{
			IDSubjectColumnField: subjID,
			VersionField:         old.Version,
		},
	)

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		return fmt.Errorf("to sql: %w", err)
	}

	res, err := s.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if affected == 0 {
		return ErrorNotFound
	}

	return nil
}
