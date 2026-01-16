package postgres

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

type PaginationFilter struct {
	Limit     int
	Asc       bool
	SortLabel string
	IDLabel   string
	LastID    *string
}

const (
	AscSortLabel  = "ASC"
	DescSortLabel = "DESC"
)

func MakeQueryWithPagination(ctx context.Context, b sq.SelectBuilder, filter *PaginationFilter) (string, []interface{}, error) {
	if filter == nil {
		return "", nil, fmt.Errorf("invalid pagination")
	}

	order := AscSortLabel
	if !filter.Asc {
		order = DescSortLabel
	}
	b = b.OrderBy(fmt.Sprintf("%s %s", filter.SortLabel, order))

	if filter.LastID != nil {
		if !filter.Asc {
			b = b.Where(sq.Lt{filter.IDLabel: filter.LastID})
		} else {
			b = b.Where(sq.Gt{filter.IDLabel: filter.LastID})
		}
	}

	query, args, err := b.
		Limit(uint64(filter.Limit)).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("build select with pagination: %w", err)
	}

	return query, args, nil
}
