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

	indexTmpl := template.Must(template.ParseFS(templates.FS, "index.html"))

	habitRepository := repository.NewSQLiteHabitRepository(conn)
	habitService := service.NewHabitService(habitRepository)

	dailyRecordRepo := repository.NewDailyRecord(conn)
	habitDoneService := service.NewHabitDone(dailyRecordRepo, nil)

	h := handler.New(indexTmpl, habitService, habitDoneService)

	mux := http.NewServeMux()
	mux.Handle("/", h)

	addr := ":8080"
	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
