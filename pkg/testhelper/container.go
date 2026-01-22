package testhelper

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer represents a running Postgres container for testing
type PostgresContainer struct {
	Container *postgres.PostgresContainer
	DSN       string
}

// SetupPostgres creates and starts a new Postgres container
func SetupPostgres(ctx context.Context) (*PostgresContainer, error) {
	dbName := "railzway_test"
	dbUser := "testuser"
	dbPassword := "testpassword"

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	return &PostgresContainer{
		Container: pgContainer,
		DSN:       connStr,
	}, nil
}

// Teardown terminates the container
func (c *PostgresContainer) Teardown(ctx context.Context) error {
	return c.Container.Terminate(ctx)
}
