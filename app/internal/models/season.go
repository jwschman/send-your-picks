package models

import (
	"time"
)

type Season struct {
	ID            string    `json:"id" db:"id"`                           // year's ULID
	Year          int       `json:"year" db:"year"`                       // which year it's for
	IsActive      bool      `json:"is_active" db:"is_active"`             // to close the season at the end
	CreatedAt     time.Time `json:"created_at" db:"created_at"`           // when the season was created
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`           // when the season was last edited
	CreatedBy     string    `json:"created_by" db:"created_by"`           // which user created the season
	NumberOfWeeks int       `json:"number_of_weeks" db:"number_of_weeks"` // how many weeks are in the season
	IsPostseason  bool      `json:"is_postseason" db:"is_postseason"`     // postseason flag
}

// SeasonParticipant represents a row in the season_participants table.
// This is the raw junction table record linking a user to a season.
type SeasonParticipant struct {
	SeasonID string    `json:"season_id" db:"season_id"` // the season being participated in
	UserID   string    `json:"user_id" db:"user_id"`     // the user participating
	JoinedAt time.Time `json:"joined_at" db:"joined_at"` // when the user was added to the season
}

// Participant is an enriched view of a season participant, including profile info.
// Used in API responses where we need to display participant details (username, avatar, etc.)
// rather than just the raw junction table data.
type Participant struct {
	UserID    string    `json:"user_id" db:"user_id"`       // user's UUID
	Username  *string   `json:"username" db:"username"`     // display name (nullable)
	Tagline   *string   `json:"tagline" db:"tagline"`       // optional tagline (nullable)
	AvatarURL *string   `json:"avatar_url" db:"avatar_url"` // profile picture URL (nullable)
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`   // when added to this season
}
