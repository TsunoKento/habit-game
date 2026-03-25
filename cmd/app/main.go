package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"habit-game/internal/db"
	"habit-game/internal/handler"
	"habit-game/internal/repository"
	"habit-game/internal/service"
	"habit-game/migrations"
	"habit-game/templates"
)

func main() {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "habit.db"
	}

	conn, err := db.Open(dsn, migrations.FS)
	if err != nil {
		log.Fatalf("database init: %v", err)
	}
	defer conn.Close()

	tmpl := template.Must(template.ParseFS(templates.FS, "index.html"))

	dailyRecordRepo := repository.NewDailyRecord(conn)
	habitDoneService := service.NewHabitDone(dailyRecordRepo, nil)

	h := handler.NewWithDependencies(tmpl, habitDoneService)

	addr := ":8080"
	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, h); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
