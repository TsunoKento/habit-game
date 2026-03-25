package repository_test

import (
	"context"
	"testing"

	"habit-game/internal/db"
	"habit-game/internal/repository"
	"habit-game/migrations"
)

func TestDailyRecordRepository_CreateAndExistsByHabitAndDate(t *testing.T) {
	conn, err := db.Open(":memory:", migrations.FS)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	repo := repository.NewDailyRecord(conn)
	ctx := context.Background()
	date := "2026-03-26"

	exists, err := repo.ExistsByHabitAndDate(ctx, 1, date)
	if err != nil {
		t.Fatalf("ExistsByHabitAndDate before create: %v", err)
	}
	if exists {
		t.Fatal("expected record to not exist before create")
	}

	if err := repo.Create(ctx, 1, date); err != nil {
		t.Fatalf("Create: %v", err)
	}

	exists, err = repo.ExistsByHabitAndDate(ctx, 1, date)
	if err != nil {
		t.Fatalf("ExistsByHabitAndDate after create: %v", err)
	}
	if !exists {
		t.Fatal("expected record to exist after create")
	}
}

func TestDailyRecordRepository_CreateRejectsDuplicateHabitAndDate(t *testing.T) {
	conn, err := db.Open(":memory:", migrations.FS)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	repo := repository.NewDailyRecord(conn)
	ctx := context.Background()
	date := "2026-03-26"

	if err := repo.Create(ctx, 1, date); err != nil {
		t.Fatalf("first Create: %v", err)
	}

	if err := repo.Create(ctx, 1, date); err == nil {
		t.Fatal("expected duplicate Create to return error")
	}
}
