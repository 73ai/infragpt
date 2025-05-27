package postgrestest

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/testcontainers/testcontainers-go/wait"
	"math/rand"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const IMAGE = "postgres:16-alpine"

func NewDB(t *testing.T) *sql.DB {
	ctx := context.Background()

	connStr := newPostgres(t, ctx)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close database: %v", err)
		}
	})

	runMigrations(t, db)

	return db
}

func newPostgres(t *testing.T, ctx context.Context) string {
	dbName := fmt.Sprintf("db-%d", rand.Intn(1000))
	dbUser := fmt.Sprintf("user-%d", rand.Intn(1000))
	dbPassword := fmt.Sprintf("password-%d", rand.Intn(1000))

	postgresContainer, err := postgres.Run(ctx,
		IMAGE,
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	t.Cleanup(func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate postgres container: %v", err)
		}
	})

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}
	return connStr
}

func runMigrations(t *testing.T, db *sql.DB) {
	// log current working directory
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}

	migrationsDir := fmt.Sprintf("%s/supporting/postgres/db/schema", dir)
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("failed to read schema directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		migrationPath := fmt.Sprintf("%s/%s", migrationsDir, file.Name())
		// read all sql files in the directory
		query, err := os.ReadFile(migrationPath)
		if err != nil {
			t.Fatalf("failed to read migration file: %v", err)
		}

		_, err = db.Exec(string(query))
		if err != nil {
			t.Fatalf("failed to run migration: %v", err)
		}
	}
}
