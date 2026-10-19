package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tikidata-analytics/berita-duka-webapp/internal/config"
	"github.com/tikidata-analytics/berita-duka-webapp/internal/db"
	"github.com/tikidata-analytics/berita-duka-webapp/internal/models"
	"github.com/tikidata-analytics/berita-duka-webapp/internal/views"
)

// Handler contains all HTTP handler functions and dependencies.
type Handler struct {
	repo *db.Repository
	cfg  *config.Config
}

// NewHandler initializes a new Handler instance.
func NewHandler(repo *db.Repository, cfg *config.Config) *Handler {
	return &Handler{
		repo: repo,
		cfg:  cfg,
	}
}

// RegisterRoutes mounts all application routes to the given ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.HomeHandler)
	mux.HandleFunc("/create", h.CreateFormHandler)
	mux.HandleFunc("/m/", h.MicrositeHandler)
	mux.HandleFunc("/api/obituaries", h.ObituariesAPIHandler)
	mux.HandleFunc("/api/purge", h.PurgeHandler)
}

// HomeHandler renders the homepage or handles root routing.
func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	obituaries, err := h.repo.ListRecent(r.Context(), 18)
	if err != nil {
		// If DB error, render empty list gracefully
		obituaries = []models.Obituary{}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := views.RenderHomePage(obituaries, h.cfg.AppBaseURL, h.cfg.SiteName)
	_, _ = w.Write([]byte(html))
}

// CreateFormHandler serves the self-service obituary submission form.
func (h *Handler) CreateFormHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := views.RenderCreateForm(h.cfg.AppBaseURL, h.cfg.SiteName)
	_, _ = w.Write([]byte(html))
}

// MicrositeHandler serves the individual public obituary page (/m/:slug).
func (h *Handler) MicrositeHandler(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/m/")
	slug = strings.TrimSpace(slug)
	if slug == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	ob, err := h.repo.GetBySlug(r.Context(), slug)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html><html lang="id"><body style="background:#0c0a09;color:#e7e5e4;font-family:sans-serif;text-align:center;padding:50px;">
			<h2>🕊️ Berita Duka Tidak Ditemukan</h2>
			<p>Halaman berita duka ini tidak ditemukan atau masa tayang telah berakhir.</p>
			<a href="%s" style="color:#f59e0b;text-decoration:none;">← Kembali ke Beranda</a>
		</body></html>`, h.cfg.AppBaseURL)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := views.RenderMicrosite(ob, h.cfg.AppBaseURL, h.cfg.SiteName)
	_, _ = w.Write([]byte(html))
}

// ObituariesAPIHandler handles GET (list) and POST (create) on /api/obituaries.
func (h *Handler) ObituariesAPIHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		obituaries, err := h.repo.ListRecent(r.Context(), 20)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		respondJSON(w, http.StatusOK, obituaries)

	case http.MethodPost:
		var req models.CreateObituaryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Format JSON tidak valid"})
			return
		}

		if err := req.Validate(); err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		deathDate, err := parseDate(req.DeathDate)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Format tanggal wafat tidak valid (gunakan YYYY-MM-DD)"})
			return
		}

		funeralTime, err := parseDateTime(req.FuneralTime)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Format waktu pemakaman tidak valid"})
			return
		}

		var birthDate *time.Time
		if strings.TrimSpace(req.BirthDate) != "" {
			if bd, err := parseDate(req.BirthDate); err == nil {
				birthDate = &bd
			}
		}

		var ceremonyTime *time.Time
		if strings.TrimSpace(req.CeremonyTime) != "" {
			if ct, err := parseDateTime(req.CeremonyTime); err == nil {
				ceremonyTime = &ct
			}
		}

		photoURL := req.PhotoURL
		if strings.TrimSpace(photoURL) == "" {
			photoURL = h.cfg.DefaultObituaryPhotoURL
		}

		slug := models.GenerateSlug(req.FullName, deathDate)
		now := time.Now()
		expiresAt := funeralTime.AddDate(0, 0, h.cfg.PurgeRetentionDays)

		ob := &models.Obituary{
			ID:                fmt.Sprintf("ob-%d", now.UnixNano()),
			Slug:              slug,
			FullName:          strings.TrimSpace(req.FullName),
			CallName:          strings.TrimSpace(req.CallName),
			Age:               req.Age,
			BirthDate:         birthDate,
			DeathDate:         deathDate,
			Religion:          strings.TrimSpace(req.Religion),
			PhotoURL:          photoURL,
			Bio:               strings.TrimSpace(req.Bio),
			FamilyMembers:     req.FamilyMembers,
			FuneralHomeName:   strings.TrimSpace(req.FuneralHomeName),
			FuneralHomeRoom:   strings.TrimSpace(req.FuneralHomeRoom),
			CeremonyTime:      ceremonyTime,
			FuneralTime:       funeralTime,
			CemeteryName:      strings.TrimSpace(req.CemeteryName),
			CemeteryMapsURL:   strings.TrimSpace(req.CemeteryMapsURL),
			ContactPhone:      strings.TrimSpace(req.ContactPhone),
			BankName:          strings.TrimSpace(req.BankName),
			BankAccountNumber: strings.TrimSpace(req.BankAccountNumber),
			BankAccountName:   strings.TrimSpace(req.BankAccountName),
			IsActive:          true,
			CreatedAt:         now,
			ExpiresAt:         expiresAt,
		}

		if err := h.repo.Insert(r.Context(), ob); err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Gagal menyimpan ke database: %v", err)})
			return
		}

		micrositeURL := fmt.Sprintf("%s/m/%s", h.cfg.AppBaseURL, ob.Slug)
		respondJSON(w, http.StatusCreated, map[string]interface{}{
			"success":       true,
			"slug":          ob.Slug,
			"microsite_url": micrositeURL,
			"expires_at":    ob.ExpiresAt,
		})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// PurgeHandler executes the automatic purging of expired obituaries.
// Accessible via GET or POST, protected by CRON_SECRET if configured.
func (h *Handler) PurgeHandler(w http.ResponseWriter, r *http.Request) {
	// If CRON_SECRET is configured, enforce authorization
	if h.cfg.CronSecret != "" {
		authHeader := r.Header.Get("Authorization")
		querySecret := r.URL.Query().Get("secret")

		bearerToken := strings.TrimPrefix(authHeader, "Bearer ")
		if bearerToken != h.cfg.CronSecret && querySecret != h.cfg.CronSecret {
			respondJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Unauthorized: invalid or missing CRON_SECRET",
			})
			return
		}
	}

	deletedCount, err := h.repo.PurgeExpired(r.Context(), h.cfg.PurgeRetentionDays)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":                 "success",
		"purged_records":         deletedCount,
		"retention_days_applied": h.cfg.PurgeRetentionDays,
		"executed_at":            time.Now().Format(time.RFC3339),
	})
}

func parseDate(val string) (time.Time, error) {
	val = strings.TrimSpace(val)
	layouts := []string{"2006-01-02", "02-01-2006", "2006/01/02", "02/01/2006"}
	for _, l := range layouts {
		if t, err := time.Parse(l, val); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date format: %s", val)
}

func parseDateTime(val string) (time.Time, error) {
	val = strings.TrimSpace(val)
	layouts := []string{
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
		time.RFC3339,
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, val); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid datetime format: %s", val)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
