package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"habit-game/internal/service"
)

type dailyRecordRepositoryMock struct {
	existsFn                func(ctx context.Context, habitID int64, date string) (bool, error)
	findDoneHabitIDsByDateFn func(ctx context.Context, date string) (map[int64]bool, error)
	createFn                func(ctx context.Context, habitID int64, date string) error
}

func (m *dailyRecordRepositoryMock) ExistsByHabitAndDate(ctx context.Context, habitID int64, date string) (bool, error) {
	return m.existsFn(ctx, habitID, date)
}

func (m *dailyRecordRepositoryMock) FindDoneHabitIDsByDate(ctx context.Context, date string) (map[int64]bool, error) {
	if m.findDoneHabitIDsByDateFn != nil {
		return m.findDoneHabitIDsByDateFn(ctx, date)
	}
	return map[int64]bool{}, nil
}

func (m *dailyRecordRepositoryMock) Create(ctx context.Context, habitID int64, date string) error {
	return m.createFn(ctx, habitID, date)
}

func TestHabitDoneService_MarkDoneCreatesRecordWhenNotYetDone(t *testing.T) {
	tokyoNow := time.Date(2026, 3, 26, 8, 30, 0, 0, time.FixedZone("JST", 9*60*60))

	var gotHabitID int64
	var gotDate string
	repo := &dailyRecordRepositoryMock{
		existsFn: func(ctx context.Context, habitID int64, date string) (bool, error) {
			gotHabitID = habitID
			gotDate = date
			return false, nil
		},
		createFn: func(ctx context.Context, habitID int64, date string) error {
			if habitID != 3 {
				t.Fatalf("Create habitID = %d, want 3", habitID)
			}
			if date != "2026-03-26" {
				t.Fatalf("Create date = %q, want 2026-03-26", date)
			}
			return nil
		},
	}

	svc := service.NewHabitDone(repo, func() time.Time { return tokyoNow })

	if err := svc.MarkDone(context.Background(), 3); err != nil {
		t.Fatalf("MarkDone: %v", err)
	}

	if gotHabitID != 3 {
		t.Fatalf("ExistsByHabitAndDate habitID = %d, want 3", gotHabitID)
	}
	if gotDate != "2026-03-26" {
		t.Fatalf("ExistsByHabitAndDate date = %q, want 2026-03-26", gotDate)
	}
}

func TestHabitDoneService_MarkDoneSkipsCreateWhenAlreadyDone(t *testing.T) {
	repo := &dailyRecordRepositoryMock{
		existsFn: func(ctx context.Context, habitID int64, date string) (bool, error) {
			return true, nil
		},
		createFn: func(ctx context.Context, habitID int64, date string) error {
			t.Fatal("Create should not be called when record already exists")
			return nil
		},
	}

	svc := service.NewHabitDone(repo, func() time.Time {
		return time.Date(2026, 3, 26, 9, 0, 0, 0, time.UTC)
	})

	if err := svc.MarkDone(context.Background(), 1); err != nil {
		t.Fatalf("MarkDone: %v", err)
	}
}

func TestHabitDoneService_DoneHabitIDsReturnsDoneSet(t *testing.T) {
	tokyoNow := time.Date(2026, 3, 26, 8, 30, 0, 0, time.FixedZone("JST", 9*60*60))

	repo := &dailyRecordRepositoryMock{
		existsFn: func(ctx context.Context, habitID int64, date string) (bool, error) { return false, nil },
		findDoneHabitIDsByDateFn: func(ctx context.Context, date string) (map[int64]bool, error) {
			if date != "2026-03-26" {
				t.Fatalf("unexpected date %q", date)
			}
			return map[int64]bool{1: true, 3: true}, nil
		},
		createFn: func(ctx context.Context, habitID int64, date string) error { return nil },
	}

	svc := service.NewHabitDone(repo, func() time.Time { return tokyoNow })

	ids, err := svc.DoneHabitIDs(context.Background())
	if err != nil {
		t.Fatalf("DoneHabitIDs: %v", err)
	}
	if !ids[1] {
		t.Error("expected habit 1 to be done")
	}
	if ids[2] {
		t.Error("expected habit 2 to not be done")
	}
	if !ids[3] {
		t.Error("expected habit 3 to be done")
	}
}

func TestHabitDoneService_MarkDoneReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("boom")
	repo := &dailyRecordRepositoryMock{
		existsFn: func(ctx context.Context, habitID int64, date string) (bool, error) {
			return false, wantErr
		},
		createFn: func(ctx context.Context, habitID int64, date string) error {
			t.Fatal("Create should not be called when exists check fails")
			return nil
		},
	}

	svc := service.NewHabitDone(repo, func() time.Time {
		return time.Date(2026, 3, 26, 9, 0, 0, 0, time.UTC)
	})

	err := svc.MarkDone(context.Background(), 1)
	if !errors.Is(err, wantErr) {
		t.Fatalf("MarkDone error = %v, want wrapped %v", err, wantErr)
	}
}
