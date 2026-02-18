package service

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"pawked.com/sendyourpicks/internal/models"
)

// get all picks for a given week
func GetWeekPicks(db *sqlx.DB, weekID string) ([]models.Pick, error) {
	var weekPicks []models.Pick

	query := `
		SELECT
			id,
			user_id,
			game_id,
			week_id,
			selected_team_id,
			is_correct,
			calculated_at
		FROM public.picks
		WHERE week_id = $1
	`
	err := db.Select(&weekPicks, query, weekID)
	if err != nil {
		return nil, err
	}

	return weekPicks, nil
}

// GetWeekGames returns all games for a given week
func GetWeekGames(db *sqlx.DB, weekID string) ([]models.Game, error) {
	var weekGames []models.Game

	query := `
		SELECT
			id,
			season_id,
			week_id,
			home_team_id,
			away_team_id,
			home_score,
			away_score,
			home_spread,
			status
		FROM public.games
		WHERE week_id = $1
	`

	if err := db.Select(&weekGames, query, weekID); err != nil {
		return nil, err
	}

	return weekGames, nil
}

// just returns the status of a week
func GetWeekStatus(db *sqlx.DB, weekID string) (string, error) {
	var weekStatus string
	query := `SELECT status FROM public.weeks WHERE id = $1`
	if err := db.Get(&weekStatus, query, weekID); err != nil {
		return "", err
	}
	return weekStatus, nil
}

// Just returns if a week exists
func WeekExists(db *sqlx.DB, weekID string) (bool, error) {

	var weekExists bool

	query := `
				SELECT EXISTS (
					SELECT 1
					FROM public.weeks
					WHERE id = $1
				);
			`
	if err := db.Get(&weekExists, query, weekID); err != nil {
		return false, err
	}
	return weekExists, nil
}

// returns true if a season exists
func SeasonExists(db *sqlx.DB, seasonID string) (bool, error) {
	var seasonExists bool
	query := `
        SELECT EXISTS (
            SELECT 1
            FROM public.seasons
            WHERE id = $1
        );
    `
	if err := db.Get(&seasonExists, query, seasonID); err != nil {
		return false, err
	}
	return seasonExists, nil
}

// returns a week with the year from the database if it exists
func GetWeekWithYear(db *sqlx.DB, weekID string) (*models.WeekWithYear, error) {
	var week models.WeekWithYear

	err := db.Get(&week, `
		SELECT
			w.id,
			w.season_id,
			w.number,
			w.status,
			s.year,
			s.is_postseason
		FROM weeks w
		JOIN seasons s ON s.id = w.season_id
		WHERE w.id = $1
	`, weekID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWeekNotFound
		}
		return nil, err
	}

	return &week, nil
}

// GetSeasonIDbyYear returns the season ID for a given year
func GetSeasonIDbyYear(db *sqlx.DB, year int) (string, error) {
	var seasonID string

	err := db.Get(
		&seasonID,
		`SELECT id FROM public.seasons WHERE year = $1`,
		year,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("season not found")
		}
		return "", err
	}
	return seasonID, nil
}

// GetTeamIDByAbbreviation looks up an active team's ID by its abbreviation
func GetTeamIDByAbbreviation(db *sqlx.DB, abbreviation string) (string, error) {
	var teamID string
	err := db.Get(&teamID,
		`SELECT id FROM public.teams WHERE abbreviation = $1 AND is_active = true`,
		abbreviation,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("team mapping failed for abbr %s: %w", abbreviation, ErrTeamMappingFailed)
		}
		return "", err
	}
	return teamID, nil
}

// UpdateWeekStatus updates a week's status and updated_at timestamp
func UpdateWeekStatus(tx *sqlx.Tx, weekID string, status string) error {
	_, err := tx.Exec(
		`UPDATE public.weeks SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, weekID,
	)
	return err
}

// IsUserSeasonParticipant checks if a user is a participant in the season
// that contains the given week. This is used to gate write operations
// like submitting or locking picks.
func IsUserSeasonParticipant(db *sqlx.DB, weekID string, userID string) (bool, error) {
	var isParticipant bool

	// Join weeks to season_participants to check membership in one query.
	// This avoids needing to first lookup the season_id separately.
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM public.season_participants sp
			JOIN public.weeks w ON w.season_id = sp.season_id
			WHERE w.id = $1 AND sp.user_id = $2
		)
	`
	if err := db.Get(&isParticipant, query, weekID, userID); err != nil {
		return false, err
	}
	return isParticipant, nil
}
