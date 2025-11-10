package storage

import (
	"auth/internal/entities"
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/huandu/go-sqlbuilder"
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
	query := sq.NewSelectBuilder()
	query.Select(SelectAllField).
		From(SubjectTableField).
		Where(
			query.Equal(IDSubjectColumnField, subjID),
			query.IsNull(DeletedAtColumnField),
		)

	sqlStr, args := query.BuildWithFlavor(sq.PostgreSQL)

	var sm SubjectModel
	err := s.db.GetContext(ctx, &sm, sqlStr, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrorNotFound
		}
		return nil, fmt.Errorf("query row context: %w", err)
	}

	return sm.ToEntity(), nil
}

func (s *Storage) CreateSubject(ctx context.Context, login, password string) (int, error) {
	query := sq.NewInsertBuilder()
	query.InsertInto(SubjectTableField).
		Cols(
			LoginColumnField,
			PasswordHashColumnField,
			CreatedAtColumnField,
		).Values(
		login,
		password,
		time.Now(),
	).Returning(
		IDSubjectColumnField,
	)

	sqlStr, args := query.BuildWithFlavor(sq.PostgreSQL)

	var id int
	err := s.db.QueryRowContext(ctx, sqlStr, args...).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("query row context: %w", err)
	}

	return id, nil
}

func (s *Storage) DeleteSubject(ctx context.Context, subjID int) error {
	query := sq.NewUpdateBuilder()

	query.Update(SubjectTableField).
		Set(
			query.Assign(DeletedAtColumnField, time.Now()),
		).Where(
		query.Equal(IDSubjectColumnField, subjID),
	)

	sqlStr, args := query.BuildWithFlavor(sq.PostgreSQL)

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

func (s *Storage) ChangePasswordHash(ctx context.Context, oldSubj *entities.Subject, newPassword string) error {
	query := sq.NewUpdateBuilder()

	query.Update(SubjectTableField).
		Set(
			query.Assign(PasswordHashColumnField, newPassword),
			query.Assign(VersionField, oldSubj.Version+1),
		).Where(
		query.Equal(IDSubjectColumnField, oldSubj.ID),
		query.Equal(VersionField, oldSubj.Version),
	)

	sqlStr, args := query.BuildWithFlavor(sq.PostgreSQL)

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
