package service_test

import (
	"context"
	"errors"
	"testing"

	"habit-game/internal/model"
	"habit-game/internal/service"
)

type recordCounterMock struct {
	countByHabitIDFn func(ctx context.Context) (map[int64]int, error)
}

func (m *recordCounterMock) CountByHabitID(ctx context.Context) (map[int64]int, error) {
	if m.countByHabitIDFn != nil {
		return m.countByHabitIDFn(ctx)
	}
	return map[int64]int{}, nil
}

var defaultHabits = []model.Habit{
	{ID: 1, Name: "早起き", ExpPerDone: 33},
	{ID: 2, Name: "英語学習", ExpPerDone: 33},
	{ID: 3, Name: "運動", ExpPerDone: 34},
}

func TestExpService_Calculate_NoRecords(t *testing.T) {
	svc := service.NewExpService(&recordCounterMock{})

	result, err := svc.Calculate(context.Background(), defaultHabits)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	if result.TotalExp != 0 {
		t.Errorf("TotalExp = %d, want 0", result.TotalExp)
	}
	if result.Level != 1 {
		t.Errorf("Level = %d, want 1", result.Level)
	}
}

func TestExpService_Calculate_WithRecords(t *testing.T) {
	mock := &recordCounterMock{
		countByHabitIDFn: func(ctx context.Context) (map[int64]int, error) {
			return map[int64]int{1: 3, 2: 1, 3: 2}, nil
		},
	}
	svc := service.NewExpService(mock)

	result, err := svc.Calculate(context.Background(), defaultHabits)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	// 33*3 + 33*1 + 34*2 = 99 + 33 + 68 = 200
	if result.TotalExp != 200 {
		t.Errorf("TotalExp = %d, want 200", result.TotalExp)
	}
	if result.Level != 3 {
		t.Errorf("Level = %d, want 3", result.Level)
	}
}

func TestExpService_Calculate_LevelBoundaries(t *testing.T) {
	singleHabit := []model.Habit{{ID: 1, Name: "test", ExpPerDone: 1}}

	tests := []struct {
		name      string
		counts    map[int64]int
		wantLevel int
	}{
		{"0 EXP → Lv1", map[int64]int{}, 1},
		{"99 EXP → Lv1", map[int64]int{1: 99}, 1},
		{"100 EXP → Lv2", map[int64]int{1: 100}, 2},
		{"200 EXP → Lv3", map[int64]int{1: 200}, 3},
		{"300 EXP → Lv4", map[int64]int{1: 300}, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &recordCounterMock{
				countByHabitIDFn: func(ctx context.Context) (map[int64]int, error) {
					return tt.counts, nil
				},
			}
			svc := service.NewExpService(mock)

			result, err := svc.Calculate(context.Background(), singleHabit)
			if err != nil {
				t.Fatalf("Calculate: %v", err)
			}
			if result.Level != tt.wantLevel {
				t.Errorf("Level = %d, want %d (TotalExp=%d)", result.Level, tt.wantLevel, result.TotalExp)
			}
		})
	}
}

func TestExpService_Calculate_RepositoryError(t *testing.T) {
	mock := &recordCounterMock{
		countByHabitIDFn: func(ctx context.Context) (map[int64]int, error) {
			return nil, errors.New("db error")
		},
	}
	svc := service.NewExpService(mock)

	_, err := svc.Calculate(context.Background(), defaultHabits)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
