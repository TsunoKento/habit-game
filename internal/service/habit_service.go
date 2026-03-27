package service

import (
	"context"
	"fmt"

	"habit-game/internal/model"
	"habit-game/internal/repository"
)

type HabitService interface {
	FindAll(ctx context.Context) ([]model.Habit, error)
}

type DefaultHabitService struct {
	repo repository.HabitRepository
}

func NewHabitService(repo repository.HabitRepository) *DefaultHabitService {
	return &DefaultHabitService{repo: repo}
}

func (s *DefaultHabitService) FindAll(ctx context.Context) ([]model.Habit, error) {
	habits, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("find all habits: %w", err)
	}

	return habits, nil
}
