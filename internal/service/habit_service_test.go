package service_test

import (
	"context"
	"errors"
	"testing"

	"habit-game/internal/model"
	"habit-game/internal/service"
)

type stubHabitRepository struct {
	habits []model.Habit
	err    error
}

func (r *stubHabitRepository) FindAll(context.Context) ([]model.Habit, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.habits, nil
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
	repo := &stubHabitRepository{err: errors.New("boom")}

	svc := service.NewHabitService(repo)

	_, err := svc.FindAll(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "find all habits: boom" {
		t.Fatalf("err = %q, want %q", err.Error(), "find all habits: boom")
	}
}
