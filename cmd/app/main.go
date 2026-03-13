package main

import (
	"html/template"
	"log"
	"net/http"

	"habit-game/internal/handler"
	"habit-game/templates"
)

func main() {
	tmpl := template.Must(template.ParseFS(templates.FS, "index.html"))

	h := handler.New(tmpl)

	addr := ":8080"
	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, h); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
