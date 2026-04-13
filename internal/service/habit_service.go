package service

import (
	"context"
	"errors"
	"fmt"

	"habit-game/internal/model"
	"habit-game/internal/repository"
)

var (
	ErrExpSumInvalid   = errors.New("total exp_per_done must equal 100")
	ErrExpValueInvalid = errors.New("each exp_per_done must be at least 1")
)

type HabitService interface {
	FindAll(ctx context.Context) ([]model.Habit, error)
	UpdateExpPerDone(ctx context.Context, updates map[int64]int) error
}

type DefaultHabitService struct {
	repo repository.HabitRepository
}

func NewHabitService(repo repository.HabitRepository) *DefaultHabitService {
	return &DefaultHabitService{repo: repo}
}

func (s *DefaultHabitService) UpdateExpPerDone(ctx context.Context, updates map[int64]int) error {
	var sum int
	for _, v := range updates {
		if v < 1 {
			return ErrExpValueInvalid
		}
		sum += v
	}
	if sum != 100 {
		return ErrExpSumInvalid
	}

	if err := s.repo.UpdateExpPerDone(ctx, updates); err != nil {
		return fmt.Errorf("update exp_per_done: %w", err)
	}
	return nil
}

func (s *DefaultHabitService) FindAll(ctx context.Context) ([]model.Habit, error) {
	habits, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("find all habits: %w", err)
	}

	return habits, nil
}
