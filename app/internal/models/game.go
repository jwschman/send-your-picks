package models

import (
	"time"
)

type Game struct {
	ID             string    `json:"id" db:"id"`
	ExternalGameID string    `json:"external_game_id" db:"external_game_id"`
	SeasonID       string    `json:"season_id" db:"season_id"`
	WeekID         string    `json:"week_id" db:"week_id"`
	HomeTeamID     string    `json:"home_team_id" db:"home_team_id"`
	AwayTeamID     string    `json:"away_team_id" db:"away_team_id"`
	HomeScore      *int      `json:"home_score" db:"home_score"`
	AwayScore      *int      `json:"away_score" db:"away_score"`
	HomeSpread     *float64  `json:"home_spread" db:"home_spread"`
	KickoffTime    time.Time `json:"kickoff_time" db:"kickoff_time"`
	NeutralSite    bool      `json:"neutral_site" db:"neutral_site"`
	Status         string    `json:"status" db:"status"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy      string    `json:"created_by" db:"created_by"`
	HomeTeamAbbr   string    `json:"home_team_abbr" db:"home_team_abbr"`
	AwayTeamAbbr   string    `json:"away_team_abbr" db:"away_team_abbr"`

	// names are populated with JOIN in query when necessary, not from the games table
	HomeTeamName    string `json:"home_team_name,omitempty" db:"home_team_name"`
	HomeTeamCity    string `json:"home_team_city,omitempty" db:"home_team_city"`
	HomeTeamLogoURL string `json:"home_team_logo_url,omitempty" db:"home_team_logo_url"`
	AwayTeamName    string `json:"away_team_name,omitempty" db:"away_team_name"`
	AwayTeamCity    string `json:"away_team_city,omitempty" db:"away_team_city"`
	AwayTeamLogoURL string `json:"away_team_logo_url,omitempty" db:"away_team_logo_url"`
}
