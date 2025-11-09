package main

import (
	"log"

	_ "github.com/lzimin05/course-todo/docs"
	"github.com/lzimin05/course-todo/config"
	"github.com/lzimin05/course-todo/internal/app"
)

// @title           Course by Leonid Zimin
// @version         1.0
// @description     API сервер для планирования задач

// @contact.name   @ZiminLeonid
// @contact.url    https://github.com/lzimin05/course-todo

// @host      localhost:8080
// @BasePath  /api

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