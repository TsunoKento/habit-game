package handler_test

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"habit-game/internal/handler"
	"habit-game/templates"
)

type habitDoneServiceStub struct {
	markDoneFn func(ctx context.Context, habitID int64) error
}

func (s habitDoneServiceStub) MarkDone(ctx context.Context, habitID int64) error {
	return s.markDoneFn(ctx, habitID)
}

func TestGetDashboard(t *testing.T) {
	tmpl := template.Must(template.New("index").Parse(`<h1>Habit Growth Tracker</h1>`))
	h := handler.New(tmpl)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Result().StatusCode)
	}
	if !strings.Contains(w.Body.String(), "Habit Growth Tracker") {
		t.Errorf("body does not contain 'Habit Growth Tracker': %s", w.Body.String())
	}
}

func TestGetDashboard_RendersHabitCards(t *testing.T) {
	tmpl := template.Must(template.ParseFS(templates.FS, "index.html"))
	h := handler.New(tmpl)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Result().StatusCode)
	}
	body := w.Body.String()
	for _, want := range []string{"早起き", "英語学習", "運動", "達成済み", "disabled", "達成する"} {
		if !strings.Contains(body, want) {
			t.Errorf("body does not contain %q", want)
		}
	}
}

func TestPostHabitDone_RedirectsToDashboard(t *testing.T) {
	tmpl := template.Must(template.ParseFS(templates.FS, "index.html"))

	var gotHabitID int64
	h := handler.NewWithDependencies(tmpl, habitDoneServiceStub{
		markDoneFn: func(ctx context.Context, habitID int64) error {
			gotHabitID = habitID
			return nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/habits/2/done", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", w.Result().StatusCode)
	}
	if gotHabitID != 2 {
		t.Fatalf("MarkDone habitID = %d, want 2", gotHabitID)
	}
	if location := w.Result().Header.Get("Location"); location != "/" {
		t.Fatalf("Location = %q, want /", location)
	}
}

func TestPostHabitDone_ReturnsBadRequestForInvalidID(t *testing.T) {
	tmpl := template.Must(template.ParseFS(templates.FS, "index.html"))
	h := handler.NewWithDependencies(tmpl, habitDoneServiceStub{
		markDoneFn: func(ctx context.Context, habitID int64) error {
			t.Fatal("MarkDone should not be called for invalid ID")
			return nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/habits/not-a-number/done", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Result().StatusCode)
	}
}
