package handler

import (
	"bytes"
	"context"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"habit-game/internal/model"
)

type Handler struct {
	tmpl            *template.Template
	habitDoneService habitDoneService
}

type habitDoneService interface {
	MarkDone(ctx context.Context, habitID int64) error
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
	return NewWithDependencies(tmpl, nil)
}

func NewWithDependencies(tmpl *template.Template, habitDoneService habitDoneService) http.Handler {
	h := &Handler{
		tmpl:            tmpl,
		habitDoneService: habitDoneService,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.handleDashboard)
	mux.HandleFunc("POST /habits/{id}/done", h.handleHabitDone)
	return mux
}

func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	data := model.DashboardData{
		Today:      formatDate(time.Now()),
		TotalLevel: 7,
		TotalExp:   210,
		Habits: []model.HabitCard{
			{ID: 1, Name: "早起き", Done: true, Level: 3, TotalExp: 90, Streak: 7},
			{ID: 2, Name: "英語学習", Done: false, Level: 2, TotalExp: 50, Streak: 3},
			{ID: 3, Name: "運動", Done: false, Level: 2, TotalExp: 70, Streak: 0},
		},
	}
	var buf bytes.Buffer
	if err := h.tmpl.Execute(&buf, data); err != nil {
		log.Printf("template render error: %v", err)
		http.Error(w, "render error", http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
}

func (h *Handler) handleHabitDone(w http.ResponseWriter, r *http.Request) {
	if h.habitDoneService == nil {
		http.Error(w, "service unavailable", http.StatusInternalServerError)
		return
	}

	habitID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid habit id", http.StatusBadRequest)
		return
	}

	if err := h.habitDoneService.MarkDone(r.Context(), habitID); err != nil {
		log.Printf("mark done error: %v", err)
		http.Error(w, "failed to record habit completion", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
