package scylla

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v3/qb"
	"github.com/scylladb/gocqlx/v3/table"
	"github.com/subipraNuvem/url-shortener/internal/domain"
)

var urlsTable = table.New(table.Metadata{
	Name:    "urls",
	Columns: []string{"code", "long_url", "is_active", "created_at", "updated_at"},
	PartKey: []string{"code"},
})

type URLRepository struct {
	client *Client
}

func NewURLRepository(client *Client) *URLRepository {
	return &URLRepository{client: client}
}

func (r *URLRepository) Create(ctx context.Context, url *domain.URL) error {
	return urlsTable.InsertQueryContext(ctx, r.client.Session()).
		BindStruct(url).
		ExecRelease()
}

func (r *URLRepository) GetByCode(ctx context.Context, code string) (*domain.URL, error) {
	var u domain.URL
	err := urlsTable.GetQueryContext(ctx, r.client.Session()).
		BindMap(qb.M{"code": code}).
		GetRelease(&u)
	if err == gocql.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get by code: %w", err)
	}
	return &u, nil
}

func (r *URLRepository) Deactivate(ctx context.Context, code string) error {
	return urlsTable.UpdateQueryContext(ctx, r.client.Session(), "is_active", "updated_at").
		BindMap(qb.M{
			"code":       code,
			"is_active":  false,
			"updated_at": time.Now().UTC(),
		}).
		ExecRelease()
}

func (r *URLRepository) IncrementClicks(ctx context.Context, code string) error {
	stmt, names := qb.Update("url_clicks").
		Add("click_count").
		Where(qb.Eq("code")).
		ToCql()
	return r.client.Session().ContextQuery(ctx, stmt, names).
		BindMap(qb.M{"click_count": 1, "code": code}).
		ExecRelease()
}

func (r *URLRepository) GetClicks(ctx context.Context, code string) (int64, error) {
	stmt, names := qb.Select("url_clicks").
		Columns("click_count").
		Where(qb.Eq("code")).
		ToCql()
	var count int64
	err := r.client.Session().ContextQuery(ctx, stmt, names).
		BindMap(qb.M{"code": code}).
		Scan(&count)
	if err == gocql.ErrNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("get clicks: %w", err)
	}
	return count, nil
}
