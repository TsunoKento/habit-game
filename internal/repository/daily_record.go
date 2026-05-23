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

func (r *DailyRecord) FindDatesByHabitID(ctx context.Context, habitID int64) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT strftime('%Y-%m-%d', date) FROM daily_records WHERE habit_id = ? ORDER BY date
	`, habitID)
	if err != nil {
		return nil, fmt.Errorf("find dates by habit: %w", err)
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			return nil, fmt.Errorf("scan date: %w", err)
		}
		dates = append(dates, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dates: %w", err)
	}
	return dates, nil
}

func (r *DailyRecord) SumExpEarned(ctx context.Context) (int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(exp_earned), 0) FROM daily_records
	`).Scan(&total); err != nil {
		return 0, fmt.Errorf("sum exp_earned: %w", err)
	}
	return total, nil
}

func (r *DailyRecord) FindByDateRange(ctx context.Context, from, to string) (map[string]map[int64]bool, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT strftime('%Y-%m-%d', date), habit_id FROM daily_records
		WHERE date BETWEEN ? AND ?
	`, from, to)
	if err != nil {
		return nil, fmt.Errorf("find by date range: %w", err)
	}
	defer rows.Close()

	result := make(map[string]map[int64]bool)
	for rows.Next() {
		var date string
		var habitID int64
		if err := rows.Scan(&date, &habitID); err != nil {
			return nil, fmt.Errorf("scan date range record: %w", err)
		}
		if result[date] == nil {
			result[date] = make(map[int64]bool)
		}
		result[date][habitID] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate date range records: %w", err)
	}
	return result, nil
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
	result, err := r.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO daily_records (habit_id, date, exp_earned)
		SELECT ?, ?, exp_per_done FROM habits WHERE id = ?
	`, habitID, date, habitID)
	if err != nil {
		return fmt.Errorf("insert daily_records: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		var exists bool
		if err := r.db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM habits WHERE id = ?)`, habitID,
		).Scan(&exists); err != nil {
			return fmt.Errorf("check habit exists: %w", err)
		}
		if !exists {
			return fmt.Errorf("habit %d not found", habitID)
		}
	}
	return nil
}
