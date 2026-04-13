package service

import (
	"context"
	"fmt"
	"time"

	"habit-game/internal/model"
)

type habitFinder interface {
	FindAll(ctx context.Context) ([]model.Habit, error)
}

type dateRangeFinder interface {
	FindByDateRange(ctx context.Context, from, to string) (map[string]map[int64]bool, error)
}

type HistoryService struct {
	habits  habitFinder
	records dateRangeFinder
	now     func() time.Time
}

func NewHistoryService(habits habitFinder, records dateRangeFinder, now func() time.Time) *HistoryService {
	if now == nil {
		now = time.Now
	}
	return &HistoryService{habits: habits, records: records, now: now}
}

func (s *HistoryService) BuildHistory(ctx context.Context, rangeType string) (*model.HistoryData, error) {
	habits, err := s.habits.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("find habits: %w", err)
	}

	today := s.now().In(JST)
	if rangeType != "week" {
		rangeType = "month"
	}

	var from time.Time
	switch rangeType {
	case "week":
		from = today.AddDate(0, 0, -6)
	case "month":
		from = time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, JST)
	}
	to := today

	fromStr := from.Format(time.DateOnly)
	toStr := to.Format(time.DateOnly)

	records, err := s.records.FindByDateRange(ctx, fromStr, toStr)
	if err != nil {
		return nil, fmt.Errorf("find records: %w", err)
	}

	var rows []model.HistoryRow
	for d := to; !d.Before(from); d = d.AddDate(0, 0, -1) {
		dateStr := d.Format(time.DateOnly)
		rows = append(rows, model.HistoryRow{
			Date:        dateStr,
			DoneByHabit: records[dateStr],
		})
	}

	return &model.HistoryData{
		Habits:       habits,
		Rows:         rows,
		CurrentRange: rangeType,
	}, nil
}
