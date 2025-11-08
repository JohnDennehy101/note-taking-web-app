package main_test

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"
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

func newTestApplication(_ *testing.T) *testApp {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	return &testApp{
		AppInterface: api.NewApplication(testDB, logger, "testing", []string{"http://localhost:3000"}),
	}
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
