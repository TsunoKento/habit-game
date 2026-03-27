package handler_test

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"habit-game/internal/handler"
	"habit-game/internal/model"
	"habit-game/templates"
)

type mockHabitService struct {
	habits []model.Habit
	err    error
}

func (m *mockHabitService) FindAll(_ context.Context) ([]model.Habit, error) {
	return m.habits, m.err
}

func TestGetDashboard(t *testing.T) {
	tmpl := template.Must(template.New("index").Parse(`<h1>Habit Growth Tracker</h1>`))
	svc := &mockHabitService{}
	h := handler.New(tmpl, svc)

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
	svc := &mockHabitService{
		habits: []model.Habit{
			{ID: 1, Name: "早起き"},
			{ID: 2, Name: "英語学習"},
			{ID: 3, Name: "運動"},
		},
	}
	h := handler.New(tmpl, svc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Result().StatusCode)
	}
	body := w.Body.String()
	for _, want := range []string{"早起き", "英語学習", "運動", "達成する"} {
		if !strings.Contains(body, want) {
			t.Errorf("body does not contain %q", want)
		}
	}
}

func TestGetDashboard_Returns500WhenServiceFails(t *testing.T) {
	tmpl := template.Must(template.ParseFS(templates.FS, "index.html"))
	svc := &mockHabitService{err: errors.New("db down")}
	h := handler.New(tmpl, svc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "internal server error") {
		t.Fatalf("unexpected body: %q", w.Body.String())
	}
}
