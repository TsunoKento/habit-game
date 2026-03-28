package repository

import (
	"context"
	"database/sql"
	"fmt"

	"habit-game/internal/model"
)

type HabitRepository interface {
	FindAll(ctx context.Context) ([]model.Habit, error)
}

type SQLiteHabitRepository struct {
	db *sql.DB
}

func NewSQLiteHabitRepository(db *sql.DB) *SQLiteHabitRepository {
	return &SQLiteHabitRepository{db: db}
}

func (r *SQLiteHabitRepository) FindAll(ctx context.Context) ([]model.Habit, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, exp_per_done, created_at
		FROM habits
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("find all habits: %w", err)
	}
	defer rows.Close()

	habits := make([]model.Habit, 0)
	for rows.Next() {
		var habit model.Habit
		if err := rows.Scan(&habit.ID, &habit.Name, &habit.ExpPerDone, &habit.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan habit: %w", err)
		}
		habits = append(habits, habit)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate habits: %w", err)
	}

	return habits, nil
}
