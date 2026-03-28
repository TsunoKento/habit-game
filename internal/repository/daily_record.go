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

func (r *DailyRecord) FindDoneHabitIDsByDate(ctx context.Context, date string) (map[int64]bool, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT habit_id FROM daily_records WHERE date = ?
	`, date)
	if err != nil {
		return nil, fmt.Errorf("find done habits: %w", err)
	}
	defer rows.Close()

	done := make(map[int64]bool)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan habit_id: %w", err)
		}
		done[id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate done habits: %w", err)
	}
	return done, nil
}

func (r *DailyRecord) DeleteByHabitAndDate(ctx context.Context, habitID int64, date string) error {
	if _, err := r.db.ExecContext(ctx, `
		DELETE FROM daily_records WHERE habit_id = ? AND date = ?
	`, habitID, date); err != nil {
		return fmt.Errorf("delete daily_records: %w", err)
	}
	return nil
}

func (r *DailyRecord) Create(ctx context.Context, habitID int64, date string) error {
	if _, err := r.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO daily_records (habit_id, date)
		VALUES (?, ?)
	`, habitID, date); err != nil {
		return fmt.Errorf("insert daily_records: %w", err)
	}
	return nil
}
