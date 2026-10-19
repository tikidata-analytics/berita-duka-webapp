package config

import (
	"bufio"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Config represents all application configuration parameters loaded from environment.
type Config struct {
	Port                   string
	AppEnv                 string
	AppBaseURL             string
	SiteName               string
	DatabaseURL            string
	AutoMigrate            bool
	MaxDBConns             int32
	MinDBConns             int32
	DBConnTimeoutSec       int
	PurgeRetentionDays     int
	CronSecret             string
	MaxUploadSizeMB        int64
	DefaultObituaryPhotoURL string

	// Cloudinary Image Storage Integration
	CloudinaryCloudName    string
	CloudinaryAPIKey       string
	CloudinaryAPISecret    string
	CloudinaryUploadPreset string
	CloudinaryFolder       string
}

// Load reads configuration from environment variables and optional .env file.
func Load() *Config {
	// Try loading .env file if it exists
	loadDotEnv(".env")

	c := &Config{
		Port:                   getEnv("PORT", "8080"),
		AppEnv:                 getEnv("APP_ENV", "development"),
		AppBaseURL:             strings.TrimRight(getEnv("APP_BASE_URL", "http://localhost:8080"), "/"),
		SiteName:               getEnv("SITE_NAME", "Portal Berita Duka & Lelayu Digital"),
		DatabaseURL:            getEnv("DATABASE_URL", ""),
		AutoMigrate:            getEnvBool("AUTO_MIGRATE", true),
		MaxDBConns:             int32(getEnvInt("MAX_DB_CONNS", 10)),
		MinDBConns:             int32(getEnvInt("MIN_DB_CONNS", 2)),
		DBConnTimeoutSec:       getEnvInt("DB_CONN_TIMEOUT_SEC", 10),
		PurgeRetentionDays:     getEnvInt("PURGE_RETENTION_DAYS", 30),
		CronSecret:             getEnv("CRON_SECRET", ""),
		MaxUploadSizeMB:        int64(getEnvInt("MAX_UPLOAD_SIZE_MB", 5)),
		DefaultObituaryPhotoURL: getEnv("DEFAULT_OBITUARY_PHOTO_URL", "https://images.unsplash.com/photo-1518495973542-4542c06a5843?auto=format&fit=crop&w=800&q=80"),

		CloudinaryCloudName:    getEnv("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryAPIKey:       getEnv("CLOUDINARY_API_KEY", ""),
		CloudinaryAPISecret:    getEnv("CLOUDINARY_API_SECRET", ""),
		CloudinaryUploadPreset: getEnv("CLOUDINARY_UPLOAD_PRESET", ""),
		CloudinaryFolder:       getEnv("CLOUDINARY_FOLDER", "berita-duka"),
	}

	// Parse unified CLOUDINARY_URL format if provided: cloudinary://<api_key>:<api_secret>@<cloud_name>
	cloudinaryURL := getEnv("CLOUDINARY_URL", "")
	if cloudinaryURL != "" && c.CloudinaryCloudName == "" {
		if parsed, err := url.Parse(cloudinaryURL); err == nil && parsed.Scheme == "cloudinary" {
			c.CloudinaryCloudName = parsed.Host
			if parsed.User != nil {
				c.CloudinaryAPIKey = parsed.User.Username()
				if pass, ok := parsed.User.Password(); ok {
					c.CloudinaryAPISecret = pass
				}
			}
		}
	}

	return c
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		lower := strings.ToLower(strings.TrimSpace(val))
		return lower == "true" || lower == "1" || lower == "yes"
	}
	return fallback
}

// loadDotEnv parses a simple key=value .env file without external dependencies.
func loadDotEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return // File does not exist, silently continue
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			v = strings.Trim(v, `"'`)
			// Only set if not already set by system environment
			if _, exists := os.LookupEnv(k); !exists {
				_ = os.Setenv(k, v)
			}
		}
	}
}
