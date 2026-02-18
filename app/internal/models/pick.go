package models

import (
	"time"
)

type Pick struct {
	ID             string     `json:"id" db:"id"`
	UserID         string     `json:"user_id" db:"user_id"`
	GameID         string     `json:"game_id" db:"game_id"`
	WeekID         string     `json:"week_id" db:"week_id"`
	SelectedTeamID *string    `json:"selected_team_id" db:"selected_team_id"` // nullable
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	IsCorrect      *bool      `json:"is_correct" db:"is_correct"`
	CalculatedAt   *time.Time `json:"calculated_at" db:"calculated_at"`
	UserLockedAt   *time.Time `json:"user_locked_at" db:"user_locked_at"`
}

type PickSummary struct {
	WeekID            string `db:"week_id" json:"week_id"`
	TotalGames        int    `db:"total_games" json:"total_games"`
	PicksCompleted    int    `db:"picks_completed" json:"picks_completed"`
	AllPicksCompleted bool   `db:"all_picks_completed" json:"all_picks_completed"`
	AllPicksLocked    bool   `db:"all_picks_locked" json:"all_picks_locked"`
}
