package handler

import (
	"bytes"
	"context"
	"errors"
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
	historyTmpl      *template.Template
	settingsTmpl     *template.Template
	service          service.HabitService
	habitDoneService habitDoneService
	expService       expService
	historyService   historyService
}

type expService interface {
	Calculate(ctx context.Context, habits []model.Habit) (*service.ExpResult, error)
}

type historyService interface {
	BuildHistory(ctx context.Context, rangeType string) (*model.HistoryData, error)
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

func New(tmpl *template.Template, historyTmpl *template.Template, settingsTmpl *template.Template, svc service.HabitService, doneSvc habitDoneService, expSvc expService, historySvc historyService) http.Handler {
	h := &Handler{
		tmpl:             tmpl,
		historyTmpl:      historyTmpl,
		settingsTmpl:     settingsTmpl,
		service:          svc,
		habitDoneService: doneSvc,
		expService:       expSvc,
		historyService:   historySvc,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.handleDashboard)
	mux.HandleFunc("GET /history", h.handleHistory)
	mux.HandleFunc("GET /settings", h.handleSettings)
	mux.HandleFunc("POST /settings", h.handleUpdateSettings)
	mux.HandleFunc("POST /habits/{id}/done", h.handleHabitDone)
	mux.HandleFunc("POST /habits/{id}/undone", h.handleHabitUndone)
	return mux
}

func (h *Handler) handleSettings(w http.ResponseWriter, r *http.Request) {
	habits, err := h.service.FindAll(r.Context())
	if err != nil {
		log.Printf("find habits error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := model.NewSettingsData(habits, nil, "")

	var buf bytes.Buffer
	if err := h.settingsTmpl.Execute(&buf, data); err != nil {
		log.Printf("template render error: %v", err)
		http.Error(w, "render error", http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
}

func (h *Handler) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	habits, err := h.service.FindAll(r.Context())
	if err != nil {
		log.Printf("find habits error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	updates := make(map[int64]int, len(habits))
	for _, hab := range habits {
		key := "exp_" + strconv.FormatInt(hab.ID, 10)
		val, err := strconv.Atoi(r.FormValue(key))
		if err != nil {
			http.Error(w, "invalid exp value", http.StatusBadRequest)
			return
		}
		updates[hab.ID] = val
	}

	if err := h.service.UpdateExpPerDone(r.Context(), updates); err != nil {
		var errMsg string
		if errors.Is(err, service.ErrExpSumInvalid) {
			errMsg = "基本経験値の合計は100にしてください"
		} else if errors.Is(err, service.ErrExpValueInvalid) {
			errMsg = "各基本経験値は1以上にしてください"
		}
		if errMsg != "" {
			data := model.NewSettingsData(habits, updates, errMsg)
			var buf bytes.Buffer
			if err := h.settingsTmpl.Execute(&buf, data); err != nil {
				log.Printf("template render error: %v", err)
				http.Error(w, "render error", http.StatusInternalServerError)
				return
			}
			buf.WriteTo(w)
			return
		}
		log.Printf("update settings error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/settings", http.StatusSeeOther)
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
			ID:     hab.ID,
			Name:   hab.Name,
			Done:   doneIDs[hab.ID],
			Streak: streak,
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

func (h *Handler) handleHistory(w http.ResponseWriter, r *http.Request) {
	rangeType := r.URL.Query().Get("range")

	data, err := h.historyService.BuildHistory(r.Context(), rangeType)
	if err != nil {
		log.Printf("build history error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := h.historyTmpl.Execute(&buf, data); err != nil {
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
