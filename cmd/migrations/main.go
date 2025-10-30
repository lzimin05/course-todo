package main

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lzimin05/course-todo/config"
	"github.com/lzimin05/course-todo/internal/infrastructure/repository"
	"github.com/pkg/errors"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	dsn, err := repository.GetConnectionString(cfg.DBConfig)
	if err != nil {
		log.Fatalf("Can't connect to database: %v", err)
	}
	
	m, err := migrate.New(
		cfg.MigrationsConfig.Path,
		dsn,
	)
	if err != nil {
		log.Panicf("Error initializing migrations: %v", err)
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("Error applying migrations: %v", err)
	}

	log.Println("Migrations applied successfully.")
}