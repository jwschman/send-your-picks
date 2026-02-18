package models

import (
	"time"
)

type Week struct {
	ID           string     `json:"id" db:"id"`
	SeasonID     string     `json:"season_id" db:"season_id"`
	Year         int        `json:"year,omitempty" db:"year"`
	IsPostseason bool       `json:"is_postseason,omitempty" db:"is_postseason"`
	Number       int        `json:"number" db:"number"` // Week 1, 2, 3, etc.
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy    string     `json:"created_by" db:"created_by"`
	Status       string     `json:"status" db:"status"`
	ActivatedAt  *time.Time `json:"activated_at" db:"activated_at"`
	ClosedAt *string `json:"closed_at" db:"closed_at"`

	// Games is optional and not in the database
	Games []Game `json:"games"`
}

type WeekWithYear struct {
	ID           string `json:"id" db:"id"`
	SeasonID     string `json:"season_id" db:"season_id"`
	Number       int    `json:"number" db:"number"`
	Status       string `json:"status" db:"status"`
	Year         int    `json:"year" db:"year"`
	IsPostseason bool   `json:"is_postseason" db:"is_postseason"`
}
