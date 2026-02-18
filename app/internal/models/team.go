package models

import "time"

type Team struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Abbreviation string    `json:"abbreviation" db:"abbreviation"`
	City         string    `json:"city" db:"city"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	LogoURL      *string   `json:"logo_url" db:"logo_url"` // Nullable?
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
