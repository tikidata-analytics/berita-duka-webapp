package handler

import (
	"context"
	"net/http"
	"sync"

	"github.com/tikidata-analytics/berita-duka-webapp/internal/config"
	"github.com/tikidata-analytics/berita-duka-webapp/internal/db"
	"github.com/tikidata-analytics/berita-duka-webapp/internal/handlers"
)

var (
	once    sync.Once
	mux     *http.ServeMux
	initErr error
)

func initApp() {
	cfg := config.Load()
	ctx := context.Background()

	var repo *db.Repository
	if cfg.DatabaseURL != "" {
		var err error
		repo, err = db.NewRepository(ctx, cfg)
		if err != nil {
			initErr = err
		}
	}

	h := handlers.NewHandler(repo, cfg)
	mux = http.NewServeMux()
	h.RegisterRoutes(mux)
}

// Handler is the serverless function entrypoint for Vercel.
func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(initApp)

	if initErr != nil && r.URL.Path != "/create" && r.URL.Path != "/" {
		http.Error(w, "Database configuration error: "+initErr.Error(), http.StatusInternalServerError)
		return
	}

	mux.ServeHTTP(w, r)
}
