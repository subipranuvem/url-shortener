package database

import "context"

type Database interface {
	Connect(ctx context.Context) error
	PingDatabasePeriodically(ctx context.Context)
	Ping(ctx context.Context) error
	Close()
}
