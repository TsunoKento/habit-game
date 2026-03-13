package handler_test

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"habit-game/internal/handler"
)

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
