package service

import (
	"context"
	"fmt"
)

type expSummer interface {
	SumExpEarned(ctx context.Context) (int, error)
}

type ExpService struct {
	summer expSummer
}

func NewExpService(summer expSummer) *ExpService {
	return &ExpService{summer: summer}
}

type ExpResult struct {
	TotalExp int
	Level    int
}

func (s *ExpService) Calculate(ctx context.Context) (*ExpResult, error) {
	totalExp, err := s.summer.SumExpEarned(ctx)
	if err != nil {
		return nil, fmt.Errorf("sum exp earned: %w", err)
	}

	return &ExpResult{
		TotalExp: totalExp,
		Level:    totalExp/100 + 1,
	}, nil
}
