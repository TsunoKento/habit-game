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
}

func (s *ExpService) Calculate(ctx context.Context, habits []model.Habit) (*ExpResult, error) {
	counts, err := s.counter.CountByHabitID(ctx)
	if err != nil {
		return nil, fmt.Errorf("count records: %w", err)
	}

	totalExp := 0
	for _, h := range habits {
		totalExp += counts[h.ID] * h.ExpPerDone
	}

	return &ExpResult{
		TotalExp: totalExp,
		Level:    totalExp/100 + 1,
	}, nil
}
