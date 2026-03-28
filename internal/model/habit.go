package model

import "time"

type Habit struct {
	ID         int
	Name       string
	ExpPerDone int
	CreatedAt  time.Time
}
