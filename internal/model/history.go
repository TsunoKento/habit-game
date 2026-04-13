package model

type HistoryRow struct {
	Date        string
	DoneByHabit map[int64]bool
}

type HistoryData struct {
	Habits       []Habit
	Rows         []HistoryRow
	CurrentRange string
}
