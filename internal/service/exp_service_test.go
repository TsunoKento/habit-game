package service_test

import (
	"context"
	"errors"
	"testing"

	"habit-game/internal/service"
)

type expSummerMock struct {
	sumFn func(ctx context.Context) (int, error)
}

func (m *expSummerMock) SumExpEarned(ctx context.Context) (int, error) {
	if m.sumFn != nil {
		return m.sumFn(ctx)
	}
	return 0, nil
}

func TestExpService_Calculate_NoRecords(t *testing.T) {
	svc := service.NewExpService(&expSummerMock{})

	result, err := svc.Calculate(context.Background())
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
	mock := &expSummerMock{
		sumFn: func(ctx context.Context) (int, error) {
			return 200, nil
		},
	}
	svc := service.NewExpService(mock)

	result, err := svc.Calculate(context.Background())
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	if result.TotalExp != 200 {
		t.Errorf("TotalExp = %d, want 200", result.TotalExp)
	}
	if result.Level != 3 {
		t.Errorf("Level = %d, want 3", result.Level)
	}
}

func TestExpService_Calculate_LevelBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		totalExp  int
		wantLevel int
	}{
		{"0 EXP → Lv1", 0, 1},
		{"99 EXP → Lv1", 99, 1},
		{"100 EXP → Lv2", 100, 2},
		{"200 EXP → Lv3", 200, 3},
		{"300 EXP → Lv4", 300, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &expSummerMock{
				sumFn: func(ctx context.Context) (int, error) {
					return tt.totalExp, nil
				},
			}
			svc := service.NewExpService(mock)

			result, err := svc.Calculate(context.Background())
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
	mock := &expSummerMock{
		sumFn: func(ctx context.Context) (int, error) {
			return 0, errors.New("db error")
		},
	}
	svc := service.NewExpService(mock)

	_, err := svc.Calculate(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
