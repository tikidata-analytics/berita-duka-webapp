package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tikidata-analytics/berita-duka-webapp/internal/config"
	"github.com/tikidata-analytics/berita-duka-webapp/internal/models"
)

// Repository defines database operations for obituaries.
type Repository struct {
	pool *pgxpool.Pool
	cfg  *config.Config
}

// NewRepository initializes a PostgreSQL connection pool with Neon support and runs auto-migration.
func NewRepository(ctx context.Context, cfg *config.Config) (*Repository, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set in environment")
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DATABASE_URL: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxDBConns
	poolConfig.MinConns = cfg.MinDBConns
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 15 * time.Minute

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.DBConnTimeoutSec)*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(timeoutCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create db connection pool: %w", err)
	}

	if err := pool.Ping(timeoutCtx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	repo := &Repository{
		pool: pool,
		cfg:  cfg,
	}

	if cfg.AutoMigrate {
		if err := repo.Migrate(ctx); err != nil {
			log.Printf("[DB Warning] Auto-migration error: %v", err)
		} else {
			log.Println("[DB] Database migration completed successfully.")
		}
	}

	return repo, nil
}

// Close releases the connection pool.
func (r *Repository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}

// Migrate creates the required obituaries table and indexes.
func (r *Repository) Migrate(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS obituaries (
		id VARCHAR(64) PRIMARY KEY,
		slug VARCHAR(180) UNIQUE NOT NULL,
		full_name VARCHAR(200) NOT NULL,
		call_name VARCHAR(100),
		age INT,
		birth_date DATE,
		death_date DATE NOT NULL,
		religion VARCHAR(60),
		photo_url TEXT,
		bio TEXT,
		family_members JSONB NOT NULL DEFAULT '[]',
		funeral_home_name VARCHAR(200),
		funeral_home_room VARCHAR(100),
		ceremony_time TIMESTAMP WITH TIME ZONE,
		funeral_time TIMESTAMP WITH TIME ZONE NOT NULL,
		cemetery_name VARCHAR(200) NOT NULL,
		cemetery_maps_url TEXT,
		contact_phone VARCHAR(60),
		bank_name VARCHAR(80),
		bank_account_number VARCHAR(80),
		bank_account_name VARCHAR(120),
		is_active BOOLEAN NOT NULL DEFAULT TRUE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_obituaries_slug ON obituaries(slug);
	CREATE INDEX IF NOT EXISTS idx_obituaries_funeral_time ON obituaries(funeral_time);
	CREATE INDEX IF NOT EXISTS idx_obituaries_expires_at ON obituaries(expires_at);
	`
	_, err := r.pool.Exec(ctx, query)
	return err
}

// Insert creates a new obituary entry.
func (r *Repository) Insert(ctx context.Context, ob *models.Obituary) error {
	famJSON, err := models.FamilyMembersToJSON(ob.FamilyMembers)
	if err != nil {
		famJSON = "[]"
	}

	query := `
	INSERT INTO obituaries (
		id, slug, full_name, call_name, age, birth_date, death_date,
		religion, photo_url, bio, family_members, funeral_home_name,
		funeral_home_room, ceremony_time, funeral_time, cemetery_name,
		cemetery_maps_url, contact_phone, bank_name, bank_account_number,
		bank_account_name, is_active, created_at, expires_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7,
		$8, $9, $10, $11, $12,
		$13, $14, $15, $16,
		$17, $18, $19, $20,
		$21, $22, $23, $24
	)
	`

	_, err = r.pool.Exec(ctx, query,
		ob.ID, ob.Slug, ob.FullName, ob.CallName, ob.Age, ob.BirthDate, ob.DeathDate,
		ob.Religion, ob.PhotoURL, ob.Bio, famJSON, ob.FuneralHomeName,
		ob.FuneralHomeRoom, ob.CeremonyTime, ob.FuneralTime, ob.CemeteryName,
		ob.CemeteryMapsURL, ob.ContactPhone, ob.BankName, ob.BankAccountNumber,
		ob.BankAccountName, ob.IsActive, ob.CreatedAt, ob.ExpiresAt,
	)
	return err
}

// GetBySlug retrieves an active obituary by its unique URL slug.
func (r *Repository) GetBySlug(ctx context.Context, slug string) (*models.Obituary, error) {
	query := `
	SELECT 
		id, slug, full_name, call_name, age, birth_date, death_date,
		religion, photo_url, bio, family_members, funeral_home_name,
		funeral_home_room, ceremony_time, funeral_time, cemetery_name,
		cemetery_maps_url, contact_phone, bank_name, bank_account_number,
		bank_account_name, is_active, created_at, expires_at
	FROM obituaries
	WHERE slug = $1 AND is_active = TRUE
	LIMIT 1
	`

	var ob models.Obituary
	var famJSON []byte

	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&ob.ID, &ob.Slug, &ob.FullName, &ob.CallName, &ob.Age, &ob.BirthDate, &ob.DeathDate,
		&ob.Religion, &ob.PhotoURL, &ob.Bio, &famJSON, &ob.FuneralHomeName,
		&ob.FuneralHomeRoom, &ob.CeremonyTime, &ob.FuneralTime, &ob.CemeteryName,
		&ob.CemeteryMapsURL, &ob.ContactPhone, &ob.BankName, &ob.BankAccountNumber,
		&ob.BankAccountName, &ob.IsActive, &ob.CreatedAt, &ob.ExpiresAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("obituari tidak ditemukan atau telah kadaluwarsa")
		}
		return nil, err
	}

	if len(famJSON) > 0 {
		_ = json.Unmarshal(famJSON, &ob.FamilyMembers)
	}

	return &ob, nil
}

// ListRecent retrieves the latest active obituaries for the home portal.
func (r *Repository) ListRecent(ctx context.Context, limit int) ([]models.Obituary, error) {
	if limit <= 0 || limit > 50 {
		limit = 12
	}

	query := `
	SELECT 
		id, slug, full_name, call_name, age, birth_date, death_date,
		religion, photo_url, bio, family_members, funeral_home_name,
		funeral_home_room, ceremony_time, funeral_time, cemetery_name,
		cemetery_maps_url, contact_phone, bank_name, bank_account_number,
		bank_account_name, is_active, created_at, expires_at
	FROM obituaries
	WHERE is_active = TRUE
	ORDER BY created_at DESC
	LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.Obituary
	for rows.Next() {
		var ob models.Obituary
		var famJSON []byte
		err := rows.Scan(
			&ob.ID, &ob.Slug, &ob.FullName, &ob.CallName, &ob.Age, &ob.BirthDate, &ob.DeathDate,
			&ob.Religion, &ob.PhotoURL, &ob.Bio, &famJSON, &ob.FuneralHomeName,
			&ob.FuneralHomeRoom, &ob.CeremonyTime, &ob.FuneralTime, &ob.CemeteryName,
			&ob.CemeteryMapsURL, &ob.ContactPhone, &ob.BankName, &ob.BankAccountNumber,
			&ob.BankAccountName, &ob.IsActive, &ob.CreatedAt, &ob.ExpiresAt,
		)
		if err == nil {
			if len(famJSON) > 0 {
				_ = json.Unmarshal(famJSON, &ob.FamilyMembers)
			}
			results = append(results, ob)
		}
	}

	return results, nil
}

// PurgeExpired deletes obituaries older than retentionDays after funeral or whose expires_at is in the past.
func (r *Repository) PurgeExpired(ctx context.Context, retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		retentionDays = 30
	}

	query := `
	DELETE FROM obituaries
	WHERE expires_at < NOW() 
	   OR funeral_time < NOW() - ($1 || ' days')::INTERVAL
	`

	res, err := r.pool.Exec(ctx, query, fmt.Sprintf("%d", retentionDays))
	if err != nil {
		return 0, err
	}

	return res.RowsAffected(), nil
}
