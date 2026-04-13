package service_test

import (
	"context"
	"testing"
	"time"

	"habit-game/internal/model"
	"habit-game/internal/service"
)

type stubHabitFinder struct {
	habits []model.Habit
	err    error
}

func (s *stubHabitFinder) FindAll(_ context.Context) ([]model.Habit, error) {
	return s.habits, s.err
}

type stubDateRangeFinder struct {
	result map[string]map[int64]bool
	err    error
}

func (s *stubDateRangeFinder) FindByDateRange(_ context.Context, _, _ string) (map[string]map[int64]bool, error) {
	return s.result, s.err
}

func fixedNow(year int, month time.Month, day int) func() time.Time {
	return func() time.Time {
		return time.Date(year, month, day, 12, 0, 0, 0, service.JST)
	}
}

func TestHistoryService_BuildHistory_Month(t *testing.T) {
	habits := []model.Habit{
		{ID: 1, Name: "早起き"},
		{ID: 2, Name: "英語学習"},
	}
	records := map[string]map[int64]bool{
		"2026-04-01": {1: true, 2: true},
		"2026-04-03": {1: true},
	}

	svc := service.NewHistoryService(
		&stubHabitFinder{habits: habits},
		&stubDateRangeFinder{result: records},
		fixedNow(2026, time.April, 5),
	)

	data, err := svc.BuildHistory(context.Background(), "month")
	if err != nil {
		t.Fatalf("BuildHistory: %v", err)
	}

	if data.CurrentRange != "month" {
		t.Errorf("CurrentRange = %q, want %q", data.CurrentRange, "month")
	}
	if len(data.Habits) != 2 {
		t.Fatalf("expected 2 habits, got %d", len(data.Habits))
	}
	// 当月1日〜今日 = 5日分の行
	if len(data.Rows) != 5 {
		t.Fatalf("expected 5 rows, got %d", len(data.Rows))
	}
	// 4/1 の habit 1 は達成
	found := false
	for _, row := range data.Rows {
		if row.Date == "2026-04-01" {
			found = true
			if !row.DoneByHabit[1] {
				t.Error("expected habit 1 done on 2026-04-01")
			}
			if !row.DoneByHabit[2] {
				t.Error("expected habit 2 done on 2026-04-01")
			}
		}
		if row.Date == "2026-04-02" {
			if row.DoneByHabit[1] {
				t.Error("expected habit 1 not done on 2026-04-02")
			}
		}
	}
	if !found {
		t.Error("2026-04-01 row not found")
	}
}

func TestHistoryService_BuildHistory_Week(t *testing.T) {
	habits := []model.Habit{
		{ID: 1, Name: "早起き"},
	}
	records := map[string]map[int64]bool{
		"2026-04-10": {1: true},
	}

	svc := service.NewHistoryService(
		&stubHabitFinder{habits: habits},
		&stubDateRangeFinder{result: records},
		fixedNow(2026, time.April, 12),
	)

	data, err := svc.BuildHistory(context.Background(), "week")
	if err != nil {
		t.Fatalf("BuildHistory: %v", err)
	}

	if data.CurrentRange != "week" {
		t.Errorf("CurrentRange = %q, want %q", data.CurrentRange, "week")
	}
	// 今日から6日前〜今日 = 7日分
	if len(data.Rows) != 7 {
		t.Fatalf("expected 7 rows, got %d", len(data.Rows))
	}
	// 最初の日付が4/12（今日、新しい順）であること
	if data.Rows[0].Date != "2026-04-12" {
		t.Errorf("first row date = %q, want %q", data.Rows[0].Date, "2026-04-12")
	}
	// 最後の日付が4/6（6日前）であること
	if data.Rows[6].Date != "2026-04-06" {
		t.Errorf("last row date = %q, want %q", data.Rows[6].Date, "2026-04-06")
	}
}

func TestHistoryService_BuildHistory_DescendingOrder(t *testing.T) {
	habits := []model.Habit{
		{ID: 1, Name: "早起き"},
	}

	svc := service.NewHistoryService(
		&stubHabitFinder{habits: habits},
		&stubDateRangeFinder{result: map[string]map[int64]bool{}},
		fixedNow(2026, time.April, 3),
	)

	data, err := svc.BuildHistory(context.Background(), "month")
	if err != nil {
		t.Fatalf("BuildHistory: %v", err)
	}

	// 3日分: 4/3, 4/2, 4/1 の順（新しい順）
	if len(data.Rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(data.Rows))
	}
	if data.Rows[0].Date != "2026-04-03" {
		t.Errorf("first row = %q, want %q", data.Rows[0].Date, "2026-04-03")
	}
	if data.Rows[2].Date != "2026-04-01" {
		t.Errorf("last row = %q, want %q", data.Rows[2].Date, "2026-04-01")
	}
}

func TestHistoryService_BuildHistory_InvalidRangeFallsBackToMonth(t *testing.T) {
	svc := service.NewHistoryService(
		&stubHabitFinder{habits: []model.Habit{{ID: 1, Name: "早起き"}}},
		&stubDateRangeFinder{result: map[string]map[int64]bool{}},
		fixedNow(2026, time.April, 5),
	)

	data, err := svc.BuildHistory(context.Background(), "invalid")
	if err != nil {
		t.Fatalf("BuildHistory: %v", err)
	}
	if data.CurrentRange != "month" {
		t.Errorf("CurrentRange = %q, want %q", data.CurrentRange, "month")
	}
	if len(data.Rows) != 5 {
		t.Fatalf("expected 5 rows (month 4/1-4/5), got %d", len(data.Rows))
	}
}
