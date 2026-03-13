package handler

import (
	"html/template"
	"net/http"
)

type Handler struct {
	tmpl *template.Template
}

// New constructs routes and returns http.Handler.
// tmpl is the parsed index template; Handler struct is ready for service injection in future issues.
func New(tmpl *template.Template) http.Handler {
	h := &Handler{tmpl: tmpl}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.handleDashboard)
	return mux
}

func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if err := h.tmpl.Execute(w, nil); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}
