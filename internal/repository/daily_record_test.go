package repository_test

import (
	"context"
	"testing"

	"habit-game/internal/db"
	"habit-game/internal/repository"
	"habit-game/migrations"
)

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

	done, err := repo.FindDoneHabitIDsByDate(ctx, date)
	if err != nil {
		t.Fatalf("FindDoneHabitIDsByDate after delete: %v", err)
	}
	if done[1] {
		t.Fatal("expected habit 1 to not be done after delete")
	}
}

func TestDailyRecordRepository_FindDatesByHabitID(t *testing.T) {
	conn, err := db.Open(":memory:", migrations.FS)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	repo := repository.NewDailyRecord(conn)
	ctx := context.Background()

	// habit 1 に3日分の記録を作成
	if err := repo.Create(ctx, 1, "2026-04-05"); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Create(ctx, 1, "2026-04-06"); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Create(ctx, 1, "2026-04-07"); err != nil {
		t.Fatalf("Create: %v", err)
	}
	// habit 2 に1日分（混ざらないことの確認）
	if err := repo.Create(ctx, 2, "2026-04-07"); err != nil {
		t.Fatalf("Create: %v", err)
	}

	dates, err := repo.FindDatesByHabitID(ctx, 1)
	if err != nil {
		t.Fatalf("FindDatesByHabitID: %v", err)
	}
	if len(dates) != 3 {
		t.Fatalf("expected 3 dates, got %d", len(dates))
	}
	// 日付昇順で返ること
	want := []string{"2026-04-05", "2026-04-06", "2026-04-07"}
	for i, d := range dates {
		if d != want[i] {
			t.Errorf("dates[%d] = %q, want %q", i, d, want[i])
		}
	}
}

func TestDailyRecordRepository_FindDatesByHabitID_Empty(t *testing.T) {
	conn, err := db.Open(":memory:", migrations.FS)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	repo := repository.NewDailyRecord(conn)
	ctx := context.Background()

	dates, err := repo.FindDatesByHabitID(ctx, 999)
	if err != nil {
		t.Fatalf("FindDatesByHabitID: %v", err)
	}
	if len(dates) != 0 {
		t.Fatalf("expected 0 dates, got %d", len(dates))
	}
}

func TestDailyRecordRepository_CountByHabitID(t *testing.T) {
	conn, err := db.Open(":memory:", migrations.FS)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	repo := repository.NewDailyRecord(conn)
	ctx := context.Background()

	// habit 1: 3 records
	for _, date := range []string{"2026-04-05", "2026-04-06", "2026-04-07"} {
		if err := repo.Create(ctx, 1, date); err != nil {
			t.Fatalf("Create habit 1 (%s): %v", date, err)
		}
	}
	// habit 2: 1 record
	if err := repo.Create(ctx, 2, "2026-04-07"); err != nil {
		t.Fatalf("Create habit 2 (2026-04-07): %v", err)
	}
	// habit 3: 0 records

	counts, err := repo.CountByHabitID(ctx)
	if err != nil {
		t.Fatalf("CountByHabitID: %v", err)
	}
	if counts[1] != 3 {
		t.Errorf("habit 1 count = %d, want 3", counts[1])
	}
	if counts[2] != 1 {
		t.Errorf("habit 2 count = %d, want 1", counts[2])
	}
	if counts[3] != 0 {
		t.Errorf("habit 3 count = %d, want 0", counts[3])
	}
}

func TestDailyRecordRepository_CountByHabitID_Empty(t *testing.T) {
	conn, err := db.Open(":memory:", migrations.FS)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	repo := repository.NewDailyRecord(conn)
	ctx := context.Background()

	counts, err := repo.CountByHabitID(ctx)
	if err != nil {
		t.Fatalf("CountByHabitID: %v", err)
	}
	if len(counts) != 0 {
		t.Errorf("expected empty map, got %v", counts)
	}
}

func TestDailyRecordRepository_CreateIsIdempotent(t *testing.T) {
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

	if err := repo.Create(ctx, 1, date); err != nil {
		t.Fatalf("duplicate Create should not return error: %v", err)
	}
}
