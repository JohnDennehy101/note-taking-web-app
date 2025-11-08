package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testDB     *sql.DB
	testDBOnce sync.Once
)

func GetTestDB() (*sql.DB, error) {
	var err error
	testDBOnce.Do(func() {
		ctx := context.Background()

		pgContainer, err := postgres.Run(ctx,
			"postgres:15-alpine",
			postgres.WithDatabase("notes_test"),
			postgres.WithUsername("testuser"),
			postgres.WithPassword("testpass"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).WithStartupTimeout(30*time.Second)),
		)
		if err != nil {
			return
		}

		connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			return
		}

		testDB, err = sql.Open("postgres", connStr)
		if err != nil {
			return
		}

		err = runMigrations(testDB)
		if err != nil {
			return
		}
	})

	return testDB, err
}

func runMigrations(db *sql.DB) error {
	migrationsDir := "../../migrations"

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}

	sort.Strings(migrationFiles)

	for _, filename := range migrationFiles {
		filepathMigration := filepath.Join(migrationsDir, filename)
		script, err := os.ReadFile(filepathMigration)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}

		if _, err := db.Exec(string(script)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}
	}

	return nil
}
