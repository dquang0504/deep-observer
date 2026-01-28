package main

import (
	"context"
	"log"
	"os"
	api "github.com/dquang0504/deep-observer/control-plane-api/internal/http"
	"github.com/dquang0504/deep-observer/control-plane-api/internal/db"
	"github.com/dquang0504/deep-observer/control-plane-api/internal/db/repo"
	"github.com/dquang0504/deep-observer/control-plane-api/internal/http/handlers"
	"github.com/dquang0504/deep-observer/control-plane-api/internal/validation"
)

func main(){
	ctx := context.Background()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == ""{
		log.Fatalf("DATABASE_URL is required")
	}

	// repo root so schema loader can find libs/schemas/*
	repoRoot := os.Getenv("REPO_ROOT")
	if repoRoot == ""{
		// assuming service is run from deep-observer/ root via compose
		repoRoot = "."
	}

	v, err := validation.LoadSchemas(repoRoot)
	if err != nil{
		log.Fatalf("load schemas: %v", err)
	}

	database, err := db.Connect(ctx, dsn)
	if err != nil{
		log.Fatalf("db connect: %v", err)
	}
	defer database.Pool.Close()

	servicesRepo := repo.NewServicesRepo(database.Pool)
	envRepo := repo.NewEnvironmentRepo(database.Pool)
	eventsRepo := repo.NewEventsRepo(database.Pool)

	eventsHandler := handlers.NewEventsHandler(v, servicesRepo, envRepo, eventsRepo)
	servicesHandler := handlers.NewServicesHandler(database.Pool)
	envHandler := handlers.NewEnvironmentsHandler(database.Pool)

	r := api.NewRouter(api.Deps{
		EventsHandler: eventsHandler,
		ServicesHandler: servicesHandler,
    	EnvironmentsHandler: envHandler,
	})

	addr := ":8090"
	log.Printf("control-plane-api listening on %s", addr)
	if err := r.Run(addr); err != nil{
		log.Fatal(err)
	}
}