package main

import (
	"flag"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/joho/godotenv"
)

func main() {
	dir := flag.String("dir", "migrations", "directory containing .cql files")
	flag.Parse()

	_ = godotenv.Load()

	hosts := strings.Split(mustEnv("SCYLLA_HOSTS"), ",")

	// Connect without a keyspace so CREATE KEYSPACE can run.
	cluster := gocql.NewCluster(hosts...)
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 30 * time.Second
	cluster.ConnectTimeout = 30 * time.Second

	session, err := cluster.CreateSession()
	if err != nil {
		slog.Error("failed to connect to scylla", "error", err)
		os.Exit(1)
	}

	slog.Info("connected to scylla", "hosts", hosts)

	files, err := filepath.Glob(filepath.Join(*dir, "*.cql"))
	if err != nil {
		session.Close()
		slog.Error("failed to list migration files", "error", err)
		os.Exit(1)
	}
	sort.Strings(files)

	for _, file := range files {
		slog.Info("running migration", "file", file)

		raw, err := os.ReadFile(file)
		if err != nil {
			session.Close()
			slog.Error("failed to read migration file", "file", file, "error", err)
			os.Exit(1)
		}

		for _, stmt := range splitStatements(string(raw)) {
			// gocql does not support USE statements — reconnect with the keyspace instead.
			if keyspace, ok := parseUse(stmt); ok {
				session.Close()
				cluster.Keyspace = keyspace
				session, err = cluster.CreateSession()
				if err != nil {
					slog.Error("failed to reconnect with keyspace", "keyspace", keyspace, "error", err)
					os.Exit(1)
				}
				slog.Info("switched keyspace", "keyspace", keyspace)
				continue
			}

			err = session.Query(stmt).Exec()
			if err != nil {
				session.Close()
				slog.Error("failed to execute statement", "file", file, "statement", stmt, "error", err)
				os.Exit(1)
			}
		}

		slog.Info("migration applied", "file", file)
	}

	session.Close()

	slog.Info("all migrations applied successfully")
}

func parseUse(stmt string) (string, bool) {
	upper := strings.ToUpper(strings.TrimSpace(stmt))
	if !strings.HasPrefix(upper, "USE ") {
		return "", false
	}
	return strings.TrimSpace(stmt[4:]), true
}

func splitStatements(content string) []string {
	parts := strings.Split(content, ";")
	stmts := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			stmts = append(stmts, p)
		}
	}
	return stmts
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("required environment variable not set", "key", key)
		os.Exit(1)
	}
	return v
}
