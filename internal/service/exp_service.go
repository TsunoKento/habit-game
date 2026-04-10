package service

import (
	"context"
	"fmt"

	"habit-game/internal/model"
)

type recordCounter interface {
	CountByHabitID(ctx context.Context) (map[int64]int, error)
}

type ExpService struct {
	counter recordCounter
}

func NewExpService(counter recordCounter) *ExpService {
	return &ExpService{counter: counter}
}

type ExpResult struct {
	TotalExp int
	Level    int
	HabitExp map[int64]int
}

func (s *ExpService) Calculate(ctx context.Context, habits []model.Habit) (*ExpResult, error) {
	counts, err := s.counter.CountByHabitID(ctx)
	if err != nil {
		return nil, fmt.Errorf("count records: %w", err)
	}

	habitExp := make(map[int64]int, len(habits))
	totalExp := 0
	for _, h := range habits {
		exp := counts[h.ID] * h.ExpPerDone
		habitExp[h.ID] = exp
		totalExp += exp
	}

	return &ExpResult{
		TotalExp: totalExp,
		Level:    totalExp/100 + 1,
		HabitExp: habitExp,
	}, nil
}
