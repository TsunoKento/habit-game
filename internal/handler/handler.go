package handler

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"habit-game/internal/model"
)

type Handler struct {
	tmpl *template.Template
}

var (
	jst      = mustLoadLocation("Asia/Tokyo")
	weekdays = [7]string{"日", "月", "火", "水", "木", "金", "土"}
)

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

func formatDate(t time.Time) string {
	t = t.In(jst)
	return t.Format("2006年01月02日") + "(" + weekdays[t.Weekday()] + ")"
}

func New(tmpl *template.Template) http.Handler {
	h := &Handler{tmpl: tmpl}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.handleDashboard)
	return mux
}

func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	data := model.DashboardData{
		Today:      formatDate(time.Now()),
		TotalLevel: 7,
		TotalExp:   210,
		Habits: []model.HabitCard{
			{Name: "早起き", Done: true, Level: 3, TotalExp: 90, Streak: 7},
			{Name: "英語学習", Done: false, Level: 2, TotalExp: 50, Streak: 3},
			{Name: "運動", Done: false, Level: 2, TotalExp: 70, Streak: 0},
		},
	}
	if err := h.tmpl.Execute(w, data); err != nil {
		log.Printf("template render error: %v", err)
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}
