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
	"habit-game/internal/service"
)

type Handler struct {
	tmpl             *template.Template
	service          service.HabitService
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

func New(tmpl *template.Template, svc service.HabitService, doneSvc habitDoneService) http.Handler {
	h := &Handler{
		tmpl:             tmpl,
		service:          svc,
		habitDoneService: doneSvc,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.handleDashboard)
	mux.HandleFunc("POST /habits/{id}/done", h.handleHabitDone)
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
