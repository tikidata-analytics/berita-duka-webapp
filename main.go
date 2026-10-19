package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tikidata-analytics/berita-duka-webapp/internal/config"
	"github.com/tikidata-analytics/berita-duka-webapp/internal/db"
	"github.com/tikidata-analytics/berita-duka-webapp/internal/handlers"
)

func main() {
	cfg := config.Load()

	log.Printf("Starting %s in %s mode...", cfg.SiteName, cfg.AppEnv)
	log.Printf("Base URL: %s | Port: %s | Purge Retention: %d days", cfg.AppBaseURL, cfg.Port, cfg.PurgeRetentionDays)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var repo *db.Repository
	var err error
	if cfg.DatabaseURL != "" {
		repo, err = db.NewRepository(ctx, cfg)
		if err != nil {
			log.Printf("[DB Notice] Could not connect to DATABASE_URL: %v", err)
			log.Println("[DB Notice] Server starting in limited mode (database operations will error until DATABASE_URL is configured).")
		} else {
			defer repo.Close()
		}
	} else {
		log.Println("[DB Notice] DATABASE_URL is empty. Set it in .env to enable Neon PostgreSQL persistence.")
	}

	h := handlers.NewHandler(repo, cfg)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	serverHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[PANIC RECOVERED] %s %s: %v", r.Method, r.URL.Path, rec)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		mux.ServeHTTP(w, r)
		log.Printf("%s %s [%s]", r.Method, r.URL.Path, time.Since(start))
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      serverHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down server gracefully...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server forced shutdown: %v", err)
		}
		cancel()
	}()

	fmt.Printf("\n=======================================================\n")
	fmt.Printf(" 🕊️  %s\n", cfg.SiteName)
	fmt.Printf(" Server running on: http://localhost:%s\n", cfg.Port)
	fmt.Printf(" Submit form      : http://localhost:%s/create\n", cfg.Port)
	fmt.Printf(" Purge endpoint   : http://localhost:%s/api/purge\n", cfg.Port)
	fmt.Printf("=======================================================\n\n")

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
