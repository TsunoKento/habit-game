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

func TestDailyRecordRepository_FindDoneHabitIDsByDate(t *testing.T) {
	conn, err := db.Open(":memory:", migrations.FS)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	repo := repository.NewDailyRecord(conn)
	ctx := context.Background()
	date := "2026-03-26"

	if err := repo.Create(ctx, 1, date); err != nil {
		t.Fatalf("Create habit 1: %v", err)
	}
	if err := repo.Create(ctx, 3, date); err != nil {
		t.Fatalf("Create habit 3: %v", err)
	}

	done, err := repo.FindDoneHabitIDsByDate(ctx, date)
	if err != nil {
		t.Fatalf("FindDoneHabitIDsByDate: %v", err)
	}
	if !done[1] {
		t.Error("expected habit 1 to be done")
	}
	if done[2] {
		t.Error("expected habit 2 to not be done")
	}
	if !done[3] {
		t.Error("expected habit 3 to be done")
	}
}

func TestDailyRecordRepository_DeleteByHabitAndDate(t *testing.T) {
	conn, err := db.Open(":memory:", migrations.FS)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	repo := repository.NewDailyRecord(conn)
	ctx := context.Background()
	date := "2026-03-26"

	if err := repo.Create(ctx, 1, date); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.DeleteByHabitAndDate(ctx, 1, date); err != nil {
		t.Fatalf("DeleteByHabitAndDate: %v", err)
	}

	exists, err := repo.ExistsByHabitAndDate(ctx, 1, date)
	if err != nil {
		t.Fatalf("ExistsByHabitAndDate after delete: %v", err)
	}
	if exists {
		t.Fatal("expected record to not exist after delete")
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
