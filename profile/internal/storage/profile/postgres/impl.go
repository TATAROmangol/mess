package postgres

import (
	"fmt"
	"profile/internal/model"

	sq "github.com/Masterminds/squirrel"
)

func (s *Storage) GetProfileFromSubjectID(subjID string) (*model.Profile, error) {
	sql, args, err := sq.
		Select(AllLabelsSelect).
		From(ProfileTable).
		Where(sq.Eq{SubjectIDLabel: subjID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select profile by subject id sql: %w", err)
	}

	var entity ProfileEntity
	err = s.db.Get(&entity, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("select profile by subject id: %w", err)
	}

	return entity.ToModel(), nil
}

