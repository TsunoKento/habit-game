package service

import (
	"context"
	"fmt"
	"time"
)

type dailyRecordRepository interface {
	FindDoneHabitIDsByDate(ctx context.Context, date string) (map[int64]bool, error)
	FindDatesByHabitID(ctx context.Context, habitID int64) ([]string, error)
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
	date := s.now().In(JST).Format(time.DateOnly)
	ids, err := s.repo.FindDoneHabitIDsByDate(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("find done habit IDs by date: %w", err)
	}
	return ids, nil
}

func (s *HabitDone) MarkDone(ctx context.Context, habitID int64) error {
	date := s.now().In(JST).Format(time.DateOnly)
	if err := s.repo.Create(ctx, habitID, date); err != nil {
		return fmt.Errorf("create daily record: %w", err)
	}
	return nil
}

func (s *HabitDone) MarkUndone(ctx context.Context, habitID int64) error {
	date := s.now().In(JST).Format(time.DateOnly)
	if err := s.repo.DeleteByHabitAndDate(ctx, habitID, date); err != nil {
		return fmt.Errorf("delete daily record: %w", err)
	}
	return nil
}

func (s *HabitDone) Streak(ctx context.Context, habitID int64) (int, error) {
	dates, err := s.repo.FindDatesByHabitID(ctx, habitID)
	if err != nil {
		return 0, fmt.Errorf("find dates for streak: %w", err)
	}

	today := s.now().In(JST).Format(time.DateOnly)

	// 日付をセットに変換
	dateSet := make(map[string]bool, len(dates))
	for _, d := range dates {
		dateSet[d] = true
	}

	// 起点: 今日のレコードがあれば今日、なければ昨日
	current, err := time.Parse(time.DateOnly, today)
	if err != nil {
		return 0, fmt.Errorf("parse today date: %w", err)
	}
	if !dateSet[today] {
		current = current.AddDate(0, 0, -1)
	}

	// 起点から遡ってカウント
	streak := 0
	for dateSet[current.Format(time.DateOnly)] {
		streak++
		current = current.AddDate(0, 0, -1)
	}

	return streak, nil
}

var JST = mustLoadLocation("Asia/Tokyo")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}
