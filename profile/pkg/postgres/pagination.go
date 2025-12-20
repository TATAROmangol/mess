package postgres

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

var (
	InvalidTokenError error = fmt.Errorf("invalid token")
)

const (
	AscSortLabel  = "ASC"
	DescSortLabel = "DESC"
)

type Keyer interface {
	Key() *string
}

type Sort struct {
	Field string `json:"field"`
	Asc   bool   `json:"asc"`
}

func NewSort(field string, asc bool) *Sort {
	return &Sort{
		Field: field,
		Asc:   asc,
	}
}

type Last struct {
	Field string  `json:"field"`
	Key   *string `json:"key"`
}

func NewLast(field string, key *string) *Last {
	return &Last{
		Field: field,
		Key:   key,
	}
}

type Pagination struct {
	Sort *Sort `json:"sort,omitempty"`
	Last *Last `json:"last,omitempty"`
	Size int   `json:"size"`
}

func NewPagination(size int, sort *Sort, last *Last) *Pagination {
	return &Pagination{
		Sort: sort,
		Last: last,
		Size: size,
	}
}

func (p *Pagination) Token() string {
	data, err := json.Marshal(p)
	if err != nil {
		return ""
	}

	return base64.RawURLEncoding.EncodeToString(data)
}

func ParsePaginationToken(token string) (*Pagination, error) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("decode pagination token: %w", err)
	}

	var p Pagination
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("%w: %w", InvalidTokenError, err)
	}

	return &p, nil
}

func MakeQueryWithPagination[T Keyer](ctx context.Context, db *sqlx.DB, b sq.SelectBuilder, p *Pagination) (*Pagination, []T, error) {
	if p == nil || p.Last == nil || p.Sort == nil {
		return nil, nil, fmt.Errorf("invalid pagination")
	}

	b = addPaginationQuery(b, p)

	query, args, err := b.
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, nil, fmt.Errorf("build select with pagination: %w", err)
	}

	var res []T
	err = db.SelectContext(ctx, &res, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("db get: %w", err)
	}

	newP := *p
	lnRes := len(res)
	if lnRes > p.Size {
		newP.Last.Key = res[p.Size-1].Key()
		lnRes -= 1
	} else {
		newP.Last.Key = nil
	}

	return &newP, res[:lnRes], nil
}

func addPaginationQuery(b sq.SelectBuilder, p *Pagination) sq.SelectBuilder {
	order := AscSortLabel
	if !p.Sort.Asc {
		order = DescSortLabel
	}
	b = b.OrderBy(fmt.Sprintf("%s %s", p.Sort.Field, order))

	if p.Last.Key != nil {
		if !p.Sort.Asc {
			b = b.Where(sq.Lt{p.Last.Field: p.Last.Key})
		} else {
			b = b.Where(sq.Gt{p.Last.Field: p.Last.Key})
		}
	}

	b = b.Limit(uint64(p.Size) + 1)

	return b
}
