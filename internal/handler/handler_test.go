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
	habits      []model.Habit
	err         error
	updateExpFn func(ctx context.Context, updates map[int64]int) error
}

func (m *mockHabitService) FindAll(_ context.Context) ([]model.Habit, error) {
	return m.habits, m.err
}

func (m *mockHabitService) UpdateExpPerDone(ctx context.Context, updates map[int64]int) error {
	if m.updateExpFn != nil {
		return m.updateExpFn(ctx, updates)
	}
	return nil
}

func TestGetSettings_ShowsHabitExpPerDone(t *testing.T) {
	indexTmpl := template.Must(template.ParseFS(templates.FS, "index.html"))
	settingsTmpl := template.Must(template.ParseFS(templates.FS, "settings.html"))
	svc := &mockHabitService{
		habits: []model.Habit{
			{ID: 1, Name: "早起き", ExpPerDone: 33},
			{ID: 2, Name: "英語学習", ExpPerDone: 33},
			{ID: 3, Name: "運動", ExpPerDone: 34},
		},
	}
	h := handler.New(indexTmpl, nil, settingsTmpl, svc, habitDoneServiceStub{}, expServiceStub{}, historyServiceStub{})

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Result().StatusCode)
	}
	body := w.Body.String()
	for _, want := range []string{"早起き", "英語学習", "運動", "33", "34"} {
		if !strings.Contains(body, want) {
			t.Errorf("body does not contain %q", want)
		}
	}
	if !strings.Contains(body, "合計: 100") {
		t.Errorf("body does not contain '合計: 100'")
	}
}

func TestPostSettings_ValidationError(t *testing.T) {
	indexTmpl := template.Must(template.ParseFS(templates.FS, "index.html"))
	settingsTmpl := template.Must(template.ParseFS(templates.FS, "settings.html"))
	svc := &mockHabitService{
		habits: []model.Habit{
			{ID: 1, Name: "早起き", ExpPerDone: 33},
			{ID: 2, Name: "英語学習", ExpPerDone: 33},
			{ID: 3, Name: "運動", ExpPerDone: 34},
		},
		updateExpFn: func(ctx context.Context, updates map[int64]int) error {
			return service.ErrExpSumInvalid
		},
	}
	h := handler.New(indexTmpl, nil, settingsTmpl, svc, habitDoneServiceStub{}, expServiceStub{}, historyServiceStub{})

	form := strings.NewReader("exp_1=50&exp_2=30&exp_3=10")
	req := httptest.NewRequest(http.MethodPost, "/settings", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Result().StatusCode)
	}
	body := w.Body.String()
	if !strings.Contains(body, "基本経験値の合計は100にしてください") {
		t.Errorf("body does not contain validation error message: %s", body)
	}
}

func TestPostSettings_UpdatesExpPerDone(t *testing.T) {
	indexTmpl := template.Must(template.ParseFS(templates.FS, "index.html"))
	settingsTmpl := template.Must(template.ParseFS(templates.FS, "settings.html"))
	var updatedExp map[int64]int
	svc := &mockHabitService{
		habits: []model.Habit{
			{ID: 1, Name: "早起き", ExpPerDone: 33},
			{ID: 2, Name: "英語学習", ExpPerDone: 33},
			{ID: 3, Name: "運動", ExpPerDone: 34},
		},
		updateExpFn: func(ctx context.Context, updates map[int64]int) error {
			updatedExp = updates
			return nil
		},
	}
	h := handler.New(indexTmpl, nil, settingsTmpl, svc, habitDoneServiceStub{}, expServiceStub{}, historyServiceStub{})

	form := strings.NewReader("exp_1=50&exp_2=30&exp_3=20")
	req := httptest.NewRequest(http.MethodPost, "/settings", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", w.Result().StatusCode)
	}
	if location := w.Result().Header.Get("Location"); location != "/settings" {
		t.Fatalf("Location = %q, want /settings", location)
	}
	if updatedExp == nil {
		t.Fatal("expected UpdateExpPerDone to be called")
	}
	if updatedExp[1] != 50 || updatedExp[2] != 30 || updatedExp[3] != 20 {
		t.Fatalf("unexpected updates: %v", updatedExp)
	}
}

func TestGetDashboard_HasSettingsLink(t *testing.T) {
	tmpl := template.Must(template.ParseFS(templates.FS, "index.html"))
	svc := &mockHabitService{}
	h := handler.New(tmpl, nil, nil, svc, habitDoneServiceStub{}, expServiceStub{}, historyServiceStub{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Result().StatusCode)
	}
	body := w.Body.String()
	if !strings.Contains(body, "/settings") {
		t.Errorf("body does not contain link to /settings")
	}
}

func TestGetDashboard(t *testing.T) {
	tmpl := template.Must(template.New("index").Parse(`<h1>Habit Growth Tracker</h1>`))
	svc := &mockHabitService{}
	h := handler.New(tmpl, nil, nil, svc, habitDoneServiceStub{}, expServiceStub{}, historyServiceStub{})

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
	h := handler.New(tmpl, nil, nil, svc, habitDoneServiceStub{}, expServiceStub{}, historyServiceStub{})

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
	h := handler.New(tmpl, nil, nil, svc, habitDoneServiceStub{}, expServiceStub{}, historyServiceStub{})

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
	h := handler.New(tmpl, nil, nil, svc, doneSvc, expServiceStub{}, historyServiceStub{})

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
	h := handler.New(tmpl, nil, nil, nil, habitDoneServiceStub{
		markDoneFn: func(ctx context.Context, habitID int64) error {
			gotHabitID = habitID
			return nil
		},
	}, expServiceStub{}, historyServiceStub{})

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
	h := handler.New(tmpl, nil, nil, nil, habitDoneServiceStub{
		markUndoneFn: func(ctx context.Context, habitID int64) error {
			gotHabitID = habitID
			return nil
		},
	}, expServiceStub{}, historyServiceStub{})

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
	h := handler.New(tmpl, nil, nil, nil, habitDoneServiceStub{
		markUndoneFn: func(ctx context.Context, habitID int64) error {
			t.Fatal("MarkUndone should not be called for invalid ID")
			return nil
		},
	}, expServiceStub{}, historyServiceStub{})

	req := httptest.NewRequest(http.MethodPost, "/habits/not-a-number/undone", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Result().StatusCode)
	}
}

type historyServiceStub struct {
	buildHistoryFn func(ctx context.Context, rangeType string) (*model.HistoryData, error)
}

func (s historyServiceStub) BuildHistory(ctx context.Context, rangeType string) (*model.HistoryData, error) {
	if s.buildHistoryFn != nil {
		return s.buildHistoryFn(ctx, rangeType)
	}
	return &model.HistoryData{CurrentRange: "month"}, nil
}

func TestGetHistory_RendersHistoryPage(t *testing.T) {
	indexTmpl := template.Must(template.ParseFS(templates.FS, "index.html"))
	historyTmpl := template.Must(template.New("history").Parse(
		`<h1>履歴</h1>{{range .Rows}}<tr><td>{{.Date}}</td></tr>{{end}}`,
	))
	svc := &mockHabitService{}
	historySvc := historyServiceStub{
		buildHistoryFn: func(ctx context.Context, rangeType string) (*model.HistoryData, error) {
			return &model.HistoryData{
				Habits: []model.Habit{{ID: 1, Name: "早起き"}},
				Rows: []model.HistoryRow{
					{Date: "2026-04-12", DoneByHabit: map[int64]bool{1: true}},
					{Date: "2026-04-11", DoneByHabit: map[int64]bool{}},
				},
				CurrentRange: "month",
			}, nil
		},
	}
	h := handler.New(indexTmpl, historyTmpl, nil, svc, habitDoneServiceStub{}, expServiceStub{}, historySvc)

	req := httptest.NewRequest(http.MethodGet, "/history", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "履歴") {
		t.Errorf("body does not contain '履歴': %s", body)
	}
	if !strings.Contains(body, "2026-04-12") {
		t.Errorf("body does not contain '2026-04-12': %s", body)
	}
}

func TestGetDashboard_HasHistoryLink(t *testing.T) {
	tmpl := template.Must(template.ParseFS(templates.FS, "index.html"))
	svc := &mockHabitService{}
	h := handler.New(tmpl, nil, nil, svc, habitDoneServiceStub{}, expServiceStub{}, historyServiceStub{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "/history") {
		t.Errorf("dashboard does not contain link to /history")
	}
}

func TestGetHistory_RendersActualTemplate(t *testing.T) {
	indexTmpl := template.Must(template.ParseFS(templates.FS, "index.html"))
	historyTmpl := template.Must(template.ParseFS(templates.FS, "history.html"))
	historySvc := historyServiceStub{
		buildHistoryFn: func(ctx context.Context, rangeType string) (*model.HistoryData, error) {
			return &model.HistoryData{
				Habits: []model.Habit{
					{ID: 1, Name: "早起き"},
					{ID: 2, Name: "英語学習"},
				},
				Rows: []model.HistoryRow{
					{Date: "2026-04-12", DoneByHabit: map[int64]bool{1: true}},
					{Date: "2026-04-11", DoneByHabit: map[int64]bool{1: true, 2: true}},
				},
				CurrentRange: "month",
			}, nil
		},
	}
	h := handler.New(indexTmpl, historyTmpl, nil, nil, habitDoneServiceStub{}, expServiceStub{}, historySvc)

	req := httptest.NewRequest(http.MethodGet, "/history", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	body := w.Body.String()

	// テーブルヘッダに習慣名
	for _, name := range []string{"早起き", "英語学習"} {
		if !strings.Contains(body, name) {
			t.Errorf("body does not contain %q", name)
		}
	}
	// 達成: ○、未達成: -
	if !strings.Contains(body, "○") {
		t.Errorf("body does not contain '○'")
	}
	if !strings.Contains(body, "-") {
		t.Errorf("body does not contain '-'")
	}
	// 切り替えリンク
	if !strings.Contains(body, "range=week") {
		t.Errorf("body does not contain link to week range")
	}
	if !strings.Contains(body, "range=month") {
		t.Errorf("body does not contain link to month range")
	}
	// ダッシュボードへの戻りリンク
	if !strings.Contains(body, "href=\"/\"") {
		t.Errorf("body does not contain link to dashboard")
	}
}

func TestGetHistory_WeekRange(t *testing.T) {
	indexTmpl := template.Must(template.ParseFS(templates.FS, "index.html"))
	historyTmpl := template.Must(template.New("history").Parse(
		`range={{.CurrentRange}}`,
	))
	var gotRange string
	historySvc := historyServiceStub{
		buildHistoryFn: func(ctx context.Context, rangeType string) (*model.HistoryData, error) {
			gotRange = rangeType
			return &model.HistoryData{CurrentRange: rangeType}, nil
		},
	}
	h := handler.New(indexTmpl, historyTmpl, nil, nil, habitDoneServiceStub{}, expServiceStub{}, historySvc)

	req := httptest.NewRequest(http.MethodGet, "/history?range=week", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if gotRange != "week" {
		t.Errorf("service received range = %q, want %q", gotRange, "week")
	}
	if !strings.Contains(w.Body.String(), "range=week") {
		t.Errorf("body does not contain 'range=week': %s", w.Body.String())
	}
}

func TestGetHistory_Returns500WhenServiceFails(t *testing.T) {
	indexTmpl := template.Must(template.ParseFS(templates.FS, "index.html"))
	historyTmpl := template.Must(template.New("history").Parse(`ok`))
	historySvc := historyServiceStub{
		buildHistoryFn: func(ctx context.Context, rangeType string) (*model.HistoryData, error) {
			return nil, errors.New("db down")
		},
	}
	h := handler.New(indexTmpl, historyTmpl, nil, nil, habitDoneServiceStub{}, expServiceStub{}, historySvc)

	req := httptest.NewRequest(http.MethodGet, "/history", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestGetHistory_Returns500WhenTemplateFails(t *testing.T) {
	indexTmpl := template.Must(template.ParseFS(templates.FS, "index.html"))
	historyTmpl := template.Must(template.New("history").Parse(`{{.NotExisting.Field}}`))
	h := handler.New(indexTmpl, historyTmpl, nil, nil, habitDoneServiceStub{}, expServiceStub{}, historyServiceStub{})

	req := httptest.NewRequest(http.MethodGet, "/history", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestPostHabitDone_ReturnsBadRequestForInvalidID(t *testing.T) {
	tmpl := template.Must(template.ParseFS(templates.FS, "index.html"))
	h := handler.New(tmpl, nil, nil, nil, habitDoneServiceStub{
		markDoneFn: func(ctx context.Context, habitID int64) error {
			t.Fatal("MarkDone should not be called for invalid ID")
			return nil
		},
	}, expServiceStub{}, historyServiceStub{})

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
	h := handler.New(tmpl, nil, nil, svc, doneSvc, expServiceStub{}, historyServiceStub{})

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
	h := handler.New(tmpl, nil, nil, svc, habitDoneServiceStub{}, expSvc, historyServiceStub{})

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
