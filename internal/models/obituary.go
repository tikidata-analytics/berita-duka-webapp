package models

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// FamilyRoleGroup represents a group of family members categorized by role (e.g., "Anak & Menantu").
type FamilyRoleGroup struct {
	Role  string   `json:"role"`
	Names []string `json:"names"`
}

// Obituary represents a single obituary announcement entity.
type Obituary struct {
	ID                 string            `json:"id"`
	Slug               string            `json:"slug"`
	FullName           string            `json:"full_name"`
	CallName           string            `json:"call_name,omitempty"`
	Age                int               `json:"age,omitempty"`
	BirthDate          *time.Time        `json:"birth_date,omitempty"`
	DeathDate          time.Time         `json:"death_date"`
	Religion           string            `json:"religion,omitempty"`
	PhotoURL           string            `json:"photo_url,omitempty"`
	Bio                string            `json:"bio,omitempty"`
	FamilyMembers      []FamilyRoleGroup `json:"family_members"`
	FuneralHomeName    string            `json:"funeral_home_name,omitempty"`
	FuneralHomeRoom    string            `json:"funeral_home_room,omitempty"`
	CeremonyTime       *time.Time        `json:"ceremony_time,omitempty"`
	FuneralTime        time.Time         `json:"funeral_time"`
	CemeteryName       string            `json:"cemetery_name"`
	CemeteryMapsURL    string            `json:"cemetery_maps_url,omitempty"`
	ContactPhone       string            `json:"contact_phone,omitempty"`
	BankName           string            `json:"bank_name,omitempty"`
	BankAccountNumber  string            `json:"bank_account_number,omitempty"`
	BankAccountName    string            `json:"bank_account_name,omitempty"`
	IsActive           bool              `json:"is_active"`
	CreatedAt          time.Time         `json:"created_at"`
	ExpiresAt          time.Time         `json:"expires_at"`
}

// CreateObituaryRequest represents the incoming submission payload from the public form.
type CreateObituaryRequest struct {
	FullName          string            `json:"full_name"`
	CallName          string            `json:"call_name"`
	Age               int               `json:"age"`
	BirthDate         string            `json:"birth_date"` // YYYY-MM-DD
	DeathDate         string            `json:"death_date"` // YYYY-MM-DD
	Religion          string            `json:"religion"`
	PhotoURL          string            `json:"photo_url"`
	Bio               string            `json:"bio"`
	FamilyMembersJSON string            `json:"family_members_json"`
	FamilyMembers     []FamilyRoleGroup `json:"family_members"`
	FuneralHomeName   string            `json:"funeral_home_name"`
	FuneralHomeRoom   string            `json:"funeral_home_room"`
	CeremonyTime      string            `json:"ceremony_time"` // RFC3339 or "2006-01-02T15:04"
	FuneralTime       string            `json:"funeral_time"`  // RFC3339 or "2006-01-02T15:04"
	CemeteryName      string            `json:"cemetery_name"`
	CemeteryMapsURL   string            `json:"cemetery_maps_url"`
	ContactPhone      string            `json:"contact_phone"`
	BankName          string            `json:"bank_name"`
	BankAccountNumber string            `json:"bank_account_number"`
	BankAccountName   string            `json:"bank_account_name"`
}

// Validate checks the required fields and formats.
func (r *CreateObituaryRequest) Validate() error {
	if strings.TrimSpace(r.FullName) == "" {
		return fmt.Errorf("nama lengkap almarhum/almarhumah wajib diisi")
	}
	if strings.TrimSpace(r.DeathDate) == "" {
		return fmt.Errorf("tanggal wafat wajib diisi")
	}
	if strings.TrimSpace(r.FuneralTime) == "" {
		return fmt.Errorf("jadwal pemakaman/kremasi wajib diisi")
	}
	if strings.TrimSpace(r.CemeteryName) == "" {
		return fmt.Errorf("lokasi pemakaman/krematorium wajib diisi")
	}
	return nil
}

// GenerateSlug creates a URL-safe unique slug based on full name, date, and random suffix.
func GenerateSlug(fullName string, deathDate time.Time) string {
	cleanName := strings.ToLower(strings.TrimSpace(fullName))
	// Replace non-alphanumeric with hyphen
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	cleanName = reg.ReplaceAllString(cleanName, "-")
	cleanName = strings.Trim(cleanName, "-")
	if cleanName == "" {
		cleanName = "obituari"
	}

	dateStr := deathDate.Format("2006-01-02")
	randomSuffix := fmt.Sprintf("%04d", time.Now().UnixNano()%10000)
	return fmt.Sprintf("%s-%s-%s", cleanName, dateStr, randomSuffix)
}

// FamilyMembersToJSON serializes FamilyMembers to JSON string for Postgres JSONB storage.
func FamilyMembersToJSON(members []FamilyRoleGroup) (string, error) {
	if len(members) == 0 {
		return "[]", nil
	}
	b, err := json.Marshal(members)
	if err != nil {
		return "[]", err
	}
	return string(b), nil
}
