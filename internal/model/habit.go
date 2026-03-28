package model

import "time"

type Habit struct {
	ID         int64
	Name       string
	ExpPerDone int
	CreatedAt  time.Time
}
