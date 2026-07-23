package scylla

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v3"
	"github.com/subipraNuvem/url-shortener/internal/config"
)

type Client struct {
	session gocqlx.Session
	cfg     *config.Config
	mu      sync.Mutex
	once    sync.Once
}

var (
	instance     *Client
	instanceOnce sync.Once
)

func NewClient(cfg *config.Config) *Client {
	instanceOnce.Do(func() {
		instance = &Client{cfg: cfg}
	})
	return instance
}

type sessionResult struct {
	session gocqlx.Session
	err     error
}

func (c *Client) Connect(ctx context.Context) error {
	var connErr error
	c.once.Do(func() {
		cluster := gocql.NewCluster(c.cfg.ScyllaHosts...)
		cluster.Keyspace = c.cfg.ScyllaKeyspace
		cluster.Consistency = parseConsistency(c.cfg.ScyllaConsistency)
		cluster.DisableInitialHostLookup = c.cfg.ScyllaDisableInitialHostLookup

		// gocql.CreateSession does not accept a context, so it can block indefinitely
		// when the keyspace does not exist (internal retry loop). We run it in a
		// goroutine and race against ctx.Done() so the caller can cancel cleanly.
		ch := make(chan sessionResult, 1)
		go func() {
			s, err := gocqlx.WrapSession(cluster.CreateSession())
			ch <- sessionResult{s, err}
		}()

		select {
		case r := <-ch:
			if r.err != nil {
				connErr = fmt.Errorf("scylla connect: %w", r.err)
				return
			}
			c.mu.Lock()
			c.session = r.session
			c.mu.Unlock()
			slog.InfoContext(ctx, "scylla connected", "hosts", c.cfg.ScyllaHosts)
		case <-ctx.Done():
			connErr = fmt.Errorf("scylla connect cancelled: keyspace %q may not exist, run 'make migrate' first", c.cfg.ScyllaKeyspace)
		}
	})
	return connErr
}

func (c *Client) Session() gocqlx.Session {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.session
}

func (c *Client) Ping(ctx context.Context) error {
	return c.session.ContextQuery(ctx, "SELECT now() FROM system.local", nil).ExecRelease()
}

func (c *Client) PingDatabasePeriodically(ctx context.Context) {
	ticker := time.NewTicker(c.cfg.PingDatabaseFrequency())
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				start := time.Now()
				err := c.Ping(ctx)
				if err != nil {
					slog.ErrorContext(ctx, "scylla ping failed", "error", err)
					continue
				}
				slog.InfoContext(ctx, "scylla ping ok", "latency_ms", time.Since(start).Milliseconds())
			}
		}
	}()
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.session.Close()
}

func parseConsistency(s string) gocql.Consistency {
	levels := map[string]gocql.Consistency{
		"ONE":          gocql.One,
		"TWO":          gocql.Two,
		"THREE":        gocql.Three,
		"QUORUM":       gocql.Quorum,
		"ALL":          gocql.All,
		"LOCAL_QUORUM": gocql.LocalQuorum,
		"LOCAL_ONE":    gocql.LocalOne,
	}
	if c, ok := levels[s]; ok {
		return c
	}
	return gocql.One
}
