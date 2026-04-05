package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"habit-game/internal/service"
)

type dailyRecordRepositoryMock struct {
	findDoneHabitIDsByDateFn func(ctx context.Context, date string) (map[int64]bool, error)
	createFn                 func(ctx context.Context, habitID int64, date string) error
	deleteFn                 func(ctx context.Context, habitID int64, date string) error
}

func (m *dailyRecordRepositoryMock) FindDoneHabitIDsByDate(ctx context.Context, date string) (map[int64]bool, error) {
	if m.findDoneHabitIDsByDateFn != nil {
		return m.findDoneHabitIDsByDateFn(ctx, date)
	}
	return map[int64]bool{}, nil
}

func (m *dailyRecordRepositoryMock) Create(ctx context.Context, habitID int64, date string) error {
	if m.createFn != nil {
		return m.createFn(ctx, habitID, date)
	}
	return nil
}

func (m *dailyRecordRepositoryMock) DeleteByHabitAndDate(ctx context.Context, habitID int64, date string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, habitID, date)
	}
	return nil
}

func TestHabitDoneService_MarkDoneCreatesRecord(t *testing.T) {
	tokyoNow := time.Date(2026, 3, 26, 8, 30, 0, 0, time.FixedZone("JST", 9*60*60))

	var gotHabitID int64
	var gotDate string
	repo := &dailyRecordRepositoryMock{
		createFn: func(ctx context.Context, habitID int64, date string) error {
			gotHabitID = habitID
			gotDate = date
			return nil
		},
	}

	svc := service.NewHabitDone(repo, func() time.Time { return tokyoNow })

	if err := svc.MarkDone(context.Background(), 3); err != nil {
		t.Fatalf("MarkDone: %v", err)
	}
	if gotHabitID != 3 {
		t.Fatalf("Create habitID = %d, want 3", gotHabitID)
	}
	if gotDate != "2026-03-26" {
		t.Fatalf("Create date = %q, want 2026-03-26", gotDate)
	}
}

func TestHabitDoneService_MarkDoneIsIdempotent(t *testing.T) {
	callCount := 0
	repo := &dailyRecordRepositoryMock{
		createFn: func(ctx context.Context, habitID int64, date string) error {
			callCount++
			return nil
		},
	}

	svc := service.NewHabitDone(repo, func() time.Time {
		return time.Date(2026, 3, 26, 9, 0, 0, 0, time.UTC)
	})

	if err := svc.MarkDone(context.Background(), 1); err != nil {
		t.Fatalf("first MarkDone: %v", err)
	}
	if err := svc.MarkDone(context.Background(), 1); err != nil {
		t.Fatalf("second MarkDone: %v", err)
	}
	if callCount != 2 {
		t.Fatalf("Create call count = %d, want 2", callCount)
	}
}

func TestHabitDoneService_DoneHabitIDsReturnsDoneSet(t *testing.T) {
	tokyoNow := time.Date(2026, 3, 26, 8, 30, 0, 0, time.FixedZone("JST", 9*60*60))

	repo := &dailyRecordRepositoryMock{
		findDoneHabitIDsByDateFn: func(ctx context.Context, date string) (map[int64]bool, error) {
			if date != "2026-03-26" {
				t.Fatalf("unexpected date %q", date)
			}
			return map[int64]bool{1: true, 3: true}, nil
		},
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

func TestHabitDoneService_MarkUndoneDeletesRecord(t *testing.T) {
	tokyoNow := time.Date(2026, 3, 26, 8, 30, 0, 0, time.FixedZone("JST", 9*60*60))

	var gotHabitID int64
	var gotDate string
	repo := &dailyRecordRepositoryMock{
		deleteFn: func(ctx context.Context, habitID int64, date string) error {
			gotHabitID = habitID
			gotDate = date
			return nil
		},
	}

	svc := service.NewHabitDone(repo, func() time.Time { return tokyoNow })

	if err := svc.MarkUndone(context.Background(), 2); err != nil {
		t.Fatalf("MarkUndone: %v", err)
	}
	if gotHabitID != 2 {
		t.Fatalf("DeleteByHabitAndDate habitID = %d, want 2", gotHabitID)
	}
	if gotDate != "2026-03-26" {
		t.Fatalf("DeleteByHabitAndDate date = %q, want 2026-03-26", gotDate)
	}
}

func TestHabitDoneService_MarkUndoneReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("delete boom")
	repo := &dailyRecordRepositoryMock{
		deleteFn: func(ctx context.Context, habitID int64, date string) error {
			return wantErr
		},
	}

	svc := service.NewHabitDone(repo, func() time.Time {
		return time.Date(2026, 3, 26, 9, 0, 0, 0, time.UTC)
	})

	err := svc.MarkUndone(context.Background(), 1)
	if !errors.Is(err, wantErr) {
		t.Fatalf("MarkUndone error = %v, want wrapped %v", err, wantErr)
	}
}

func TestHabitDoneService_MarkDoneReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("boom")
	repo := &dailyRecordRepositoryMock{
		createFn: func(ctx context.Context, habitID int64, date string) error {
			return wantErr
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
