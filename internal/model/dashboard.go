package model

type HabitCard struct {
	Name     string
	Done     bool
	Level    int
	TotalExp int
	Streak   int
}

type DashboardData struct {
	Today      string
	TotalLevel int
	TotalExp   int
	Habits     []HabitCard
}
