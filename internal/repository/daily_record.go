package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type DailyRecord struct {
	db *sql.DB
}

func NewDailyRecord(db *sql.DB) *DailyRecord {
	return &DailyRecord{db: db}
}

func (r *DailyRecord) ExistsByHabitAndDate(ctx context.Context, habitID int64, date string) (bool, error) {
	var exists bool
	if err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM daily_records
			WHERE habit_id = ? AND date = ?
		)
	`, habitID, date).Scan(&exists); err != nil {
		return false, fmt.Errorf("exists daily_records: %w", err)
	}
	return exists, nil
}

func (r *DailyRecord) Create(ctx context.Context, habitID int64, date string) error {
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO daily_records (habit_id, date)
		VALUES (?, ?)
	`, habitID, date); err != nil {
		return fmt.Errorf("insert daily_records: %w", err)
	}
	return nil
}
