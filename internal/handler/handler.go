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
	expService       expService
}

type expService interface {
	Calculate(ctx context.Context, habits []model.Habit) (*service.ExpResult, error)
}

type habitDoneService interface {
	MarkDone(ctx context.Context, habitID int64) error
	MarkUndone(ctx context.Context, habitID int64) error
	DoneHabitIDs(ctx context.Context) (map[int64]bool, error)
	Streak(ctx context.Context, habitID int64) (int, error)
}

var weekdays = [7]string{"日", "月", "火", "水", "木", "金", "土"}

func formatDate(t time.Time) string {
	t = t.In(service.JST)
	return t.Format("2006年01月02日") + "(" + weekdays[t.Weekday()] + ")"
}

func New(tmpl *template.Template, svc service.HabitService, doneSvc habitDoneService, expSvc expService) http.Handler {
	h := &Handler{
		tmpl:             tmpl,
		service:          svc,
		habitDoneService: doneSvc,
		expService:       expSvc,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.handleDashboard)
	mux.HandleFunc("POST /habits/{id}/done", h.handleHabitDone)
	mux.HandleFunc("POST /habits/{id}/undone", h.handleHabitUndone)
	return mux
}

func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	habits, err := h.service.FindAll(r.Context())
	if err != nil {
		log.Printf("find habits error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	doneIDs, err := h.habitDoneService.DoneHabitIDs(r.Context())
	if err != nil {
		log.Printf("fetch done habits error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	expResult, err := h.expService.Calculate(r.Context(), habits)
	if err != nil {
		log.Printf("calculate exp error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	cards := make([]model.HabitCard, len(habits))
	for i, hab := range habits {
		streak, err := h.habitDoneService.Streak(r.Context(), hab.ID)
		if err != nil {
			log.Printf("calculate streak error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		cards[i] = model.HabitCard{
			ID:       hab.ID,
			Name:     hab.Name,
			Done:     doneIDs[hab.ID],
			TotalExp: expResult.HabitExp[hab.ID],
			Streak:   streak,
		}
	}

	data := model.DashboardData{
		Today:      formatDate(time.Now()),
		TotalLevel: expResult.Level,
		TotalExp:   expResult.TotalExp,
		Habits:     cards,
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

func (h *Handler) handleHabitUndone(w http.ResponseWriter, r *http.Request) {
	habitID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid habit id", http.StatusBadRequest)
		return
	}

	if err := h.habitDoneService.MarkUndone(r.Context(), habitID); err != nil {
		log.Printf("mark undone error: %v", err)
		http.Error(w, "failed to cancel habit completion", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
