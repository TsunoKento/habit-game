package handler

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"time"

	"habit-game/internal/model"
	"habit-game/internal/service"
)

type Handler struct {
	tmpl    *template.Template
	service service.HabitService
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

func New(tmpl *template.Template, svc service.HabitService) http.Handler {
	h := &Handler{tmpl: tmpl, service: svc}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.handleDashboard)
	return mux
}

func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	habits, err := h.service.FindAll(r.Context())
	if err != nil {
		log.Printf("find habits error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	cards := make([]model.HabitCard, len(habits))
	for i, hab := range habits {
		cards[i] = model.HabitCard{Name: hab.Name}
	}

	data := model.DashboardData{
		Today:  formatDate(time.Now()),
		Habits: cards,
	}
	var buf bytes.Buffer
	if err := h.tmpl.Execute(&buf, data); err != nil {
		log.Printf("template render error: %v", err)
		http.Error(w, "render error", http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
}
