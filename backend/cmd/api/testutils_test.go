package main_test

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	api "github.com/johndennehy101/note-taking-web-app/backend/cmd/api"
	"github.com/johndennehy101/note-taking-web-app/backend/internal/data"
	"github.com/johndennehy101/note-taking-web-app/backend/internal/testutil"
	_ "github.com/lib/pq"
)

type testApp struct {
	api.AppInterface
}

func (app *testApp) routes() http.Handler {
	return app.GetRoutes()
}

var testDB *sql.DB

func TestMain(m *testing.M) {
	db, err := testutil.GetTestDB()
	if err != nil {
		panic(err)
	}

	testDB = db

	code := m.Run()

	os.Exit(code)
}

func runMigrations(db *sql.DB) {
	migrationsDir := "../../migrations"

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		panic(err)
	}

	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}

	sort.Strings(migrationFiles)

	for _, filename := range migrationFiles {
		filepath := filepath.Join(migrationsDir, filename)
		script, err := os.ReadFile(filepath)
		if err != nil {
			panic(fmt.Errorf("failed to read migration %s: %w", filename, err))
		}

		if _, err := db.Exec(string(script)); err != nil {
			panic(fmt.Errorf("failed to execute migration %s: %w", filename, err))
		}
	}
}

func newTestApplication(t *testing.T) *testApp {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	return &testApp{
		AppInterface: api.NewApplication(testDB, logger, "testing", []string{"http://localhost:3000"}),
	}
}

func newTestDB(t *testing.T) *sql.DB {
	return testDB
}

func createTestNote(t *testing.T, app *testApp, title, body string, tags []string) *data.Note {
	note := &data.Note{
		Title: title,
		Body:  body,
		Tags:  tags,
	}

	err := app.GetModels().Notes.Insert(note)
	if err != nil {
		t.Fatal(err)
	}

	return note
}
