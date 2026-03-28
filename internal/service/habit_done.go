package service

import (
	"context"
	"fmt"
	"time"
)

type dailyRecordRepository interface {
	FindDoneHabitIDsByDate(ctx context.Context, date string) (map[int64]bool, error)
	Create(ctx context.Context, habitID int64, date string) error
	DeleteByHabitAndDate(ctx context.Context, habitID int64, date string) error
}

type HabitDone struct {
	repo dailyRecordRepository
	now  func() time.Time
}

func NewHabitDone(repo dailyRecordRepository, now func() time.Time) *HabitDone {
	if now == nil {
		now = time.Now
	}
	return &HabitDone{
		repo: repo,
		now:  now,
	}
}

func (s *HabitDone) DoneHabitIDs(ctx context.Context) (map[int64]bool, error) {
	date := s.now().In(jst).Format(time.DateOnly)
	ids, err := s.repo.FindDoneHabitIDsByDate(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("find done habit IDs by date: %w", err)
	}
	return ids, nil
}

func (s *HabitDone) MarkDone(ctx context.Context, habitID int64) error {
	date := s.now().In(jst).Format(time.DateOnly)
	if err := s.repo.Create(ctx, habitID, date); err != nil {
		return fmt.Errorf("create daily record: %w", err)
	}
	return nil
}

func (s *HabitDone) MarkUndone(ctx context.Context, habitID int64) error {
	date := s.now().In(jst).Format(time.DateOnly)
	if err := s.repo.DeleteByHabitAndDate(ctx, habitID, date); err != nil {
		return fmt.Errorf("delete daily record: %w", err)
	}
	return nil
}

var jst = mustLoadLocation("Asia/Tokyo")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}
