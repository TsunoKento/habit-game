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

func TestDailyRecordRepository_FindByDateRange(t *testing.T) {
	conn, err := db.Open(":memory:", migrations.FS)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	repo := repository.NewDailyRecord(conn)
	ctx := context.Background()

	// 複数日・複数習慣の記録を作成
	if err := repo.Create(ctx, 1, "2026-04-01"); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Create(ctx, 2, "2026-04-01"); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Create(ctx, 1, "2026-04-03"); err != nil {
		t.Fatalf("Create: %v", err)
	}
	// 範囲外の記録
	if err := repo.Create(ctx, 1, "2026-03-31"); err != nil {
		t.Fatalf("Create: %v", err)
	}

	result, err := repo.FindByDateRange(ctx, "2026-04-01", "2026-04-03")
	if err != nil {
		t.Fatalf("FindByDateRange: %v", err)
	}

	// 2026-04-01 に habit 1, 2 が達成
	if !result["2026-04-01"][1] {
		t.Error("expected habit 1 done on 2026-04-01")
	}
	if !result["2026-04-01"][2] {
		t.Error("expected habit 2 done on 2026-04-01")
	}
	// 2026-04-03 に habit 1 が達成
	if !result["2026-04-03"][1] {
		t.Error("expected habit 1 done on 2026-04-03")
	}
	// 範囲外の 2026-03-31 は含まれない
	if _, ok := result["2026-03-31"]; ok {
		t.Error("expected 2026-03-31 to not be in result")
	}
}

func TestDailyRecordRepository_FindByDateRange_Empty(t *testing.T) {
	conn, err := db.Open(":memory:", migrations.FS)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	repo := repository.NewDailyRecord(conn)
	ctx := context.Background()

	result, err := repo.FindByDateRange(ctx, "2026-04-01", "2026-04-30")
	if err != nil {
		t.Fatalf("FindByDateRange: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestDailyRecordRepository_Create_SnapshotsExpPerDone(t *testing.T) {
	conn, err := db.Open(":memory:", migrations.FS)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	repo := repository.NewDailyRecord(conn)
	ctx := context.Background()

	// seed habit 1 has exp_per_done = 33 (from initial migrations)
	if err := repo.Create(ctx, 1, "2026-05-01"); err != nil {
		t.Fatalf("Create: %v", err)
	}

	var got int
	if err := conn.QueryRowContext(ctx,
		`SELECT exp_earned FROM daily_records WHERE habit_id = ? AND date = ?`,
		1, "2026-05-01",
	).Scan(&got); err != nil {
		t.Fatalf("query exp_earned: %v", err)
	}
	if got != 33 {
		t.Fatalf("exp_earned = %d, want 33 (habit 1's current exp_per_done)", got)
	}
}

func TestDailyRecordRepository_Create_ExpEarnedNotAffectedByLaterUpdate(t *testing.T) {
	conn, err := db.Open(":memory:", migrations.FS)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	dailyRepo := repository.NewDailyRecord(conn)
	habitRepo := repository.NewSQLiteHabitRepository(conn)
	ctx := context.Background()

	// habit 1 の現在の exp_per_done=33 で達成記録を作成
	if err := dailyRepo.Create(ctx, 1, "2026-05-01"); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// habit 1 の exp_per_done を 50 に変更（残りも合計100に調整）
	if err := habitRepo.UpdateExpPerDone(ctx, map[int64]int{1: 50, 2: 25, 3: 25}); err != nil {
		t.Fatalf("UpdateExpPerDone: %v", err)
	}

	var got int
	if err := conn.QueryRowContext(ctx,
		`SELECT exp_earned FROM daily_records WHERE habit_id = ? AND date = ?`,
		1, "2026-05-01",
	).Scan(&got); err != nil {
		t.Fatalf("query exp_earned: %v", err)
	}
	if got != 33 {
		t.Fatalf("exp_earned = %d, want 33 (snapshot must not be affected by later UpdateExpPerDone)", got)
	}
}

func TestDailyRecordRepository_SumExpEarned(t *testing.T) {
	conn, err := db.Open(":memory:", migrations.FS)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	repo := repository.NewDailyRecord(conn)
	ctx := context.Background()

	// habit 1 (exp 33) x 2 days, habit 3 (exp 34) x 1 day → 33*2 + 34 = 100
	for _, d := range []string{"2026-04-05", "2026-04-06"} {
		if err := repo.Create(ctx, 1, d); err != nil {
			t.Fatalf("Create habit 1 %s: %v", d, err)
		}
	}
	if err := repo.Create(ctx, 3, "2026-04-06"); err != nil {
		t.Fatalf("Create habit 3: %v", err)
	}

	got, err := repo.SumExpEarned(ctx)
	if err != nil {
		t.Fatalf("SumExpEarned: %v", err)
	}
	if got != 100 {
		t.Fatalf("SumExpEarned = %d, want 100", got)
	}
}

func TestDailyRecordRepository_SumExpEarned_Empty(t *testing.T) {
	conn, err := db.Open(":memory:", migrations.FS)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	repo := repository.NewDailyRecord(conn)
	got, err := repo.SumExpEarned(context.Background())
	if err != nil {
		t.Fatalf("SumExpEarned: %v", err)
	}
	if got != 0 {
		t.Fatalf("SumExpEarned = %d, want 0", got)
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

func TestDailyRecordRepository_Create_ReturnsErrorForUnknownHabit(t *testing.T) {
	conn, err := db.Open(":memory:", migrations.FS)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	repo := repository.NewDailyRecord(conn)
	if err := repo.Create(context.Background(), 999, "2026-05-01"); err == nil {
		t.Fatal("expected error for unknown habit_id, got nil")
	}
}
