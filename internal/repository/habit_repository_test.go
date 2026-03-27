package repository_test

import (
	"context"
	"testing"
	"time"

	"habit-game/internal/db"
	"habit-game/internal/repository"
	"habit-game/migrations"
)

func TestSQLiteHabitRepository_FindAll(t *testing.T) {
	conn, err := db.Open(":memory:", migrations.FS)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	repo := repository.NewSQLiteHabitRepository(conn)

	habits, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}

	if len(habits) != 3 {
		t.Fatalf("expected 3 habits, got %d", len(habits))
	}

	tests := []struct {
		id         int
		name       string
		expPerDone int
	}{
		{1, "早起き", 10},
		{2, "英語学習", 10},
		{3, "運動", 10},
	}

	for i, tt := range tests {
		habit := habits[i]
		if habit.ID != tt.id {
			t.Errorf("habits[%d].ID = %d, want %d", i, habit.ID, tt.id)
		}
		if habit.Name != tt.name {
			t.Errorf("habits[%d].Name = %q, want %q", i, habit.Name, tt.name)
		}
		if habit.ExpPerDone != tt.expPerDone {
			t.Errorf("habits[%d].ExpPerDone = %d, want %d", i, habit.ExpPerDone, tt.expPerDone)
		}
		if habit.CreatedAt.IsZero() {
			t.Errorf("habits[%d].CreatedAt is zero", i)
		}
	}
}

func TestSQLiteHabitRepository_FindAll_OrdersByID(t *testing.T) {
	conn, err := db.Open(":memory:", migrations.FS)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	if _, err := conn.Exec(`UPDATE habits SET created_at = ? WHERE id = 1`, time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)); err != nil {
		t.Fatalf("update created_at: %v", err)
	}

	repo := repository.NewSQLiteHabitRepository(conn)

	habits, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}

	if habits[0].ID != 1 || habits[1].ID != 2 || habits[2].ID != 3 {
		t.Fatalf("unexpected order: %#v", habits)
	}
	if !habits[0].CreatedAt.Equal(time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)) {
		t.Fatalf("habits[0].CreatedAt = %v, want %v", habits[0].CreatedAt, time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC))
	}
}
