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
	"habit-game/internal/service"
	"habit-game/templates"
)

type habitDoneServiceStub struct {
	markDoneFn     func(ctx context.Context, habitID int64) error
	markUndoneFn   func(ctx context.Context, habitID int64) error
	doneHabitIDsFn func(ctx context.Context) (map[int64]bool, error)
	streakFn       func(ctx context.Context, habitID int64) (int, error)
}

func (s habitDoneServiceStub) MarkDone(ctx context.Context, habitID int64) error {
	if s.markDoneFn != nil {
		return s.markDoneFn(ctx, habitID)
	}
	return nil
}

func (s habitDoneServiceStub) MarkUndone(ctx context.Context, habitID int64) error {
	if s.markUndoneFn != nil {
		return s.markUndoneFn(ctx, habitID)
	}
	return nil
}

func (s habitDoneServiceStub) Streak(ctx context.Context, habitID int64) (int, error) {
	if s.streakFn != nil {
		return s.streakFn(ctx, habitID)
	}
	return 0, nil
}

func (s habitDoneServiceStub) DoneHabitIDs(ctx context.Context) (map[int64]bool, error) {
	if s.doneHabitIDsFn != nil {
		return s.doneHabitIDsFn(ctx)
	}
	return map[int64]bool{}, nil
}

type expServiceStub struct {
	calculateFn func(ctx context.Context, habits []model.Habit) (*service.ExpResult, error)
}

func (s expServiceStub) Calculate(ctx context.Context, habits []model.Habit) (*service.ExpResult, error) {
	if s.calculateFn != nil {
		return s.calculateFn(ctx, habits)
	}
	return &service.ExpResult{Level: 1, HabitExp: map[int64]int{}}, nil
}

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
	h := handler.New(tmpl, svc, habitDoneServiceStub{}, expServiceStub{})

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
	h := handler.New(tmpl, svc, habitDoneServiceStub{}, expServiceStub{})

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
	h := handler.New(tmpl, svc, habitDoneServiceStub{}, expServiceStub{})

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

func TestGetDashboard_ShowsDoneStateFromService(t *testing.T) {
	tmpl := template.Must(template.ParseFS(templates.FS, "index.html"))
	svc := &mockHabitService{
		habits: []model.Habit{
			{ID: 1, Name: "早起き"},
			{ID: 2, Name: "英語学習"},
		},
	}
	doneSvc := habitDoneServiceStub{
		doneHabitIDsFn: func(ctx context.Context) (map[int64]bool, error) {
			return map[int64]bool{1: true}, nil
		},
	}
	h := handler.New(tmpl, svc, doneSvc, expServiceStub{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Result().StatusCode)
	}
	body := w.Body.String()
	if !strings.Contains(body, "取り消す") {
		t.Errorf("body does not contain '取り消す': %s", body)
	}
	if !strings.Contains(body, "達成する") {
		t.Errorf("body does not contain '達成する': %s", body)
	}
}

func TestPostHabitDone_RedirectsToDashboard(t *testing.T) {
	tmpl := template.Must(template.ParseFS(templates.FS, "index.html"))

	var gotHabitID int64
	h := handler.New(tmpl, nil, habitDoneServiceStub{
		markDoneFn: func(ctx context.Context, habitID int64) error {
			gotHabitID = habitID
			return nil
		},
	}, expServiceStub{})

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

func TestPostHabitUndone_RedirectsToDashboard(t *testing.T) {
	tmpl := template.Must(template.ParseFS(templates.FS, "index.html"))

	var gotHabitID int64
	h := handler.New(tmpl, nil, habitDoneServiceStub{
		markUndoneFn: func(ctx context.Context, habitID int64) error {
			gotHabitID = habitID
			return nil
		},
	}, expServiceStub{})

	req := httptest.NewRequest(http.MethodPost, "/habits/3/undone", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", w.Result().StatusCode)
	}
	if gotHabitID != 3 {
		t.Fatalf("MarkUndone habitID = %d, want 3", gotHabitID)
	}
	if location := w.Result().Header.Get("Location"); location != "/" {
		t.Fatalf("Location = %q, want /", location)
	}
}

func TestPostHabitUndone_ReturnsBadRequestForInvalidID(t *testing.T) {
	tmpl := template.Must(template.ParseFS(templates.FS, "index.html"))
	h := handler.New(tmpl, nil, habitDoneServiceStub{
		markUndoneFn: func(ctx context.Context, habitID int64) error {
			t.Fatal("MarkUndone should not be called for invalid ID")
			return nil
		},
	}, expServiceStub{})

	req := httptest.NewRequest(http.MethodPost, "/habits/not-a-number/undone", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Result().StatusCode)
	}
}

func TestPostHabitDone_ReturnsBadRequestForInvalidID(t *testing.T) {
	tmpl := template.Must(template.ParseFS(templates.FS, "index.html"))
	h := handler.New(tmpl, nil, habitDoneServiceStub{
		markDoneFn: func(ctx context.Context, habitID int64) error {
			t.Fatal("MarkDone should not be called for invalid ID")
			return nil
		},
	}, expServiceStub{})

	req := httptest.NewRequest(http.MethodPost, "/habits/not-a-number/done", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Result().StatusCode)
	}
}

func TestGetDashboard_ShowsStreak(t *testing.T) {
	tmpl := template.Must(template.ParseFS(templates.FS, "index.html"))
	svc := &mockHabitService{
		habits: []model.Habit{
			{ID: 1, Name: "早起き"},
			{ID: 2, Name: "英語学習"},
		},
	}
	doneSvc := habitDoneServiceStub{
		doneHabitIDsFn: func(ctx context.Context) (map[int64]bool, error) {
			return map[int64]bool{1: true}, nil
		},
		streakFn: func(ctx context.Context, habitID int64) (int, error) {
			if habitID == 1 {
				return 5, nil
			}
			return 0, nil
		},
	}
	h := handler.New(tmpl, svc, doneSvc, expServiceStub{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Result().StatusCode)
	}
	body := w.Body.String()
	if !strings.Contains(body, "5日連続") {
		t.Errorf("body does not contain '5日連続': %s", body)
	}
	if !strings.Contains(body, "0日連続") {
		t.Errorf("body does not contain '0日連続': %s", body)
	}
}

func TestGetDashboard_ShowsExpAndLevel(t *testing.T) {
	tmpl := template.Must(template.ParseFS(templates.FS, "index.html"))
	svc := &mockHabitService{
		habits: []model.Habit{
			{ID: 1, Name: "早起き", ExpPerDone: 33},
			{ID: 2, Name: "英語学習", ExpPerDone: 33},
			{ID: 3, Name: "運動", ExpPerDone: 34},
		},
	}
	expSvc := expServiceStub{
		calculateFn: func(ctx context.Context, habits []model.Habit) (*service.ExpResult, error) {
			return &service.ExpResult{
				TotalExp: 200,
				Level:    3,
				HabitExp: map[int64]int{1: 99, 2: 33, 3: 68},
			}, nil
		},
	}
	h := handler.New(tmpl, svc, habitDoneServiceStub{}, expSvc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Result().StatusCode)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Lv.3") {
		t.Errorf("body does not contain 'Lv.3'")
	}
	if !strings.Contains(body, "200 EXP") {
		t.Errorf("body does not contain '200 EXP'")
	}
	if !strings.Contains(body, "99 EXP") {
		t.Errorf("body does not contain '99 EXP' for habit 1")
	}
}
