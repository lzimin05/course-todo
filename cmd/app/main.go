package main

import (
	"log"

	"github.com/lzimin05/course-todo/config"
	"github.com/lzimin05/course-todo/internal/app"
	_ "github.com/lib/pq"
)

func main() {
	conf, err := config.NewConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	application, err := app.NewApp(conf)
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}

	application.Run()
}