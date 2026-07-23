package domain

import "time"

type URL struct {
	Code      string    `db:"code"`
	LongURL   string    `db:"long_url"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
