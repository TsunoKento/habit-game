package db_test

import (
	"testing"
	"testing/fstest"

	"habit-game/internal/db"
)

func TestOpen_CreatesTablesFromMigrations(t *testing.T) {
	migrations := fstest.MapFS{
		"001_initial_schema.sql": {
			Data: []byte(`CREATE TABLE IF NOT EXISTS habits (
				id          INTEGER PRIMARY KEY AUTOINCREMENT,
				name        TEXT NOT NULL,
				exp_per_done INTEGER NOT NULL,
				created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE TABLE IF NOT EXISTS daily_records (
				id       INTEGER PRIMARY KEY AUTOINCREMENT,
				habit_id INTEGER NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
				date     DATE NOT NULL,
				created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(habit_id, date)
			);`),
		},
	}

	conn, err := db.Open(":memory:", migrations)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	for _, table := range []string{"habits", "daily_records", "schema_migrations"} {
		var name string
		err := conn.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}

func TestOpen_SkipsAlreadyAppliedMigrations(t *testing.T) {
	const dsn = "file:testskip?mode=memory&cache=shared"

	migrations1 := fstest.MapFS{
		"001_create.sql": {
			Data: []byte(`CREATE TABLE IF NOT EXISTS habits (id INTEGER PRIMARY KEY, name TEXT NOT NULL);`),
		},
	}

	conn, err := db.Open(dsn, migrations1)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	defer conn.Close()

	// Apply migrations a second time on the same shared in-memory DB.
	// 001_create.sql should be skipped; only 002_noop.sql should be applied.
	migrations2 := fstest.MapFS{
		"001_create.sql": {
			Data: []byte(`CREATE TABLE IF NOT EXISTS habits (id INTEGER PRIMARY KEY, name TEXT NOT NULL);`),
		},
		"002_noop.sql": {
			Data: []byte(`CREATE TABLE IF NOT EXISTS noop (id INTEGER PRIMARY KEY);`),
		},
	}

	conn2, err := db.Open(dsn, migrations2)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	defer conn2.Close()

	var count int
	if err := conn2.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 migration records, got %d", count)
	}
}

func TestOpen_CloseIsIdempotent(t *testing.T) {
	conn, err := db.Open(":memory:", fstest.MapFS{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	conn.Close()
	// Calling Close twice should not panic
	conn.Close()
}
