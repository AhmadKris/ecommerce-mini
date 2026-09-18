//go:build integration

// Package testdb spins up a throwaway Postgres container (via
// testcontainers-go) migrated to the real schema, for integration tests
// across internal/repository and internal/service that need actual
// Postgres semantics (row locking, unique/foreign-key constraint errors) a
// mock can't reproduce. Only compiled with -tags=integration since it
// requires a working Docker daemon — never part of the default `go test
// ./...` run.
package testdb

import (
	"context"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// migrationsPath is relative to this package's directory
// (internal/testdb/), so it resolves the same way regardless of which
// package's test binary imports New.
const migrationsPath = "file://../../migrations"

// New starts a fresh Postgres container, applies every migration in
// migrations/ against it — the real schema, not a hand-maintained test
// fixture that could drift from it — and returns a connected *gorm.DB.
// The container is terminated automatically via t.Cleanup.
func New(t *testing.T) *gorm.DB {
	t.Helper()
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("ecommerce_mini_test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Logf("terminate postgres container: %v", err)
		}
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("get connection string: %v", err)
	}

	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		t.Fatalf("open migrate: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("apply migrations: %v", err)
	}

	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect gorm: %v", err)
	}
	return db
}
