package model

type SettingsData struct {
	Habits   []Habit
	TotalExp int
	Error    string
}

func NewSettingsData(habits []Habit, expOverrides map[int64]int, errMsg string) SettingsData {
	out := make([]Habit, len(habits))
	var total int
	for i, h := range habits {
		exp := h.ExpPerDone
		if v, ok := expOverrides[h.ID]; ok {
			exp = v
		}
		out[i] = Habit{
			ID:         h.ID,
			Name:       h.Name,
			ExpPerDone: exp,
		}
		total += exp
	}
	return SettingsData{
		Habits:   out,
		TotalExp: total,
		Error:    errMsg,
	}
}
