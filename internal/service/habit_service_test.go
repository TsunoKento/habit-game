package service_test

import (
	"context"
	"errors"
	"testing"

	"habit-game/internal/model"
	"habit-game/internal/service"
)

type stubHabitRepository struct {
	habits     []model.Habit
	err        error
	updatedExp map[int64]int
}

func (r *stubHabitRepository) FindAll(context.Context) ([]model.Habit, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.habits, nil
}

func (r *stubHabitRepository) UpdateExpPerDone(_ context.Context, updates map[int64]int) error {
	r.updatedExp = updates
	return nil
}

func TestHabitService_UpdateExpPerDone_Success(t *testing.T) {
	repo := &stubHabitRepository{}
	svc := service.NewHabitService(repo)

	updates := map[int64]int{1: 50, 2: 30, 3: 20}
	err := svc.UpdateExpPerDone(context.Background(), updates)
	if err != nil {
		t.Fatalf("UpdateExpPerDone: %v", err)
	}

	if repo.updatedExp == nil {
		t.Fatal("expected repository UpdateExpPerDone to be called")
	}
	if repo.updatedExp[1] != 50 || repo.updatedExp[2] != 30 || repo.updatedExp[3] != 20 {
		t.Fatalf("unexpected updates: %v", repo.updatedExp)
	}
}

func TestHabitService_UpdateExpPerDone_ValidationError(t *testing.T) {
	repo := &stubHabitRepository{}
	svc := service.NewHabitService(repo)

	updates := map[int64]int{1: 50, 2: 30, 3: 10} // sum = 90, not 100
	err := svc.UpdateExpPerDone(context.Background(), updates)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !errors.Is(err, service.ErrExpSumInvalid) {
		t.Fatalf("expected ErrExpSumInvalid, got %v", err)
	}
	if repo.updatedExp != nil {
		t.Fatal("repository should not be called when validation fails")
	}
}

func TestHabitService_UpdateExpPerDone_NegativeValue(t *testing.T) {
	repo := &stubHabitRepository{}
	svc := service.NewHabitService(repo)

	updates := map[int64]int{1: -50, 2: 100, 3: 50} // sum = 100 but negative value
	err := svc.UpdateExpPerDone(context.Background(), updates)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !errors.Is(err, service.ErrExpValueInvalid) {
		t.Fatalf("expected ErrExpValueInvalid, got %v", err)
	}
	if repo.updatedExp != nil {
		t.Fatal("repository should not be called when validation fails")
	}
}

func TestHabitService_FindAll(t *testing.T) {
	repo := &stubHabitRepository{
		habits: []model.Habit{
			{ID: 1, Name: "早起き", ExpPerDone: 10},
			{ID: 2, Name: "英語学習", ExpPerDone: 10},
		},
	}

	svc := service.NewHabitService(repo)

	habits, err := svc.FindAll(context.Background())
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}

	if len(habits) != 2 {
		t.Fatalf("expected 2 habits, got %d", len(habits))
	}
	if habits[0].Name != "早起き" {
		t.Fatalf("habits[0].Name = %q, want %q", habits[0].Name, "早起き")
	}
	if habits[1].Name != "英語学習" {
		t.Fatalf("habits[1].Name = %q, want %q", habits[1].Name, "英語学習")
	}
}

func TestHabitService_FindAll_ReturnsWrappedError(t *testing.T) {
	rootErr := errors.New("boom")
	repo := &stubHabitRepository{err: rootErr}

	svc := service.NewHabitService(repo)

	_, err := svc.FindAll(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, rootErr) {
		t.Fatalf("expected wrapped root error, got %v", err)
	}
	if err.Error() != "find all habits: boom" {
		t.Fatalf("err = %q, want %q", err.Error(), "find all habits: boom")
	}
}
