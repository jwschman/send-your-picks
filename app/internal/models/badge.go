package models

type BadgeType string

const (
	BadgePreviousWeekWinner   BadgeType = "previous_week_winner"
	BadgePreviousSeasonWinner BadgeType = "previous_season_winner"
	BadgePreviousSeasonLoser  BadgeType = "previous_season_loser"
)

type Badge struct {
	Type  BadgeType `json:"type"`
	Label string    `json:"label"`
}
