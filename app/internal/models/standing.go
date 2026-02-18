package models

import (
	"time"
)

type WeekResult struct {
	ID         string    `json:"id" db:"id"`
	UserID     string    `json:"user_id" db:"user_id"`
	WeekID     string    `json:"week_id" db:"week_id"`
	Points     int       `json:"points" db:"points"`
	Rank       int       `json:"rank" db:"rank"`
	ComputedAt time.Time `json:"computed_at" db:"computed_at"`
}

type SeasonStanding struct {
	ID         string    `json:"id" db:"id"`
	UserID     string    `json:"user_id" db:"user_id"`
	SeasonID   string    `json:"season_id" db:"season_id"`
	WeekID     string    `json:"week_id" db:"week_id"`
	Points     int       `json:"points" db:"points"`
	Rank       int       `json:"rank" db:"rank"`
	ComputedAt time.Time `json:"computed_at" db:"computed_at"`
}
