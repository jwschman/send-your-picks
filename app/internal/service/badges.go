package service

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"pawked.com/sendyourpicks/internal/models"
)

// GetBadges returns a map of user ID to badges for the given season.
// Badges include: previous week winner (most recent final week in this season),
// and previous season winner/loser (final standings of the prior year's season).
func GetBadges(db *sqlx.DB, seasonID string) (map[string][]models.Badge, error) {
	badges := make(map[string][]models.Badge)

	// get the season's year so we can look up the previous season
	var seasonYear int
	err := db.Get(&seasonYear, `SELECT year FROM seasons WHERE id = $1`, seasonID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return badges, nil
		}
		return nil, err
	}

	// previous week winner or winners: rank 1 in week_results for the most recent final week
	type weekWinner struct {
		UserID     string `db:"user_id"`
		WeekNumber int    `db:"week_number"`
	}

	var weekWinners []weekWinner

	err = db.Select(&weekWinners, `
		SELECT wr.user_id, w.number AS week_number
		FROM week_results wr
		JOIN weeks w ON w.id = wr.week_id
		WHERE w.season_id = $1
		AND w.status = 'final'
		AND w.number = (
			SELECT MAX(w2.number) FROM weeks w2 WHERE w2.season_id = $1 AND w2.status = 'final'
		)
		AND wr.rank = 1
	`, seasonID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	for _, w := range weekWinners {
		badges[w.UserID] = append(badges[w.UserID], models.Badge{
			Type:  models.BadgePreviousWeekWinner,
			Label: fmt.Sprintf("Week %d Winner", w.WeekNumber),
		})
	}

	// previous season winner and loser: final standings from the prior year
	var prevSeasonID string
	err = db.Get(&prevSeasonID, `SELECT id FROM seasons WHERE year = $1 AND is_postseason = false`, seasonYear-1)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return badges, nil
		}
		return nil, err
	}

	type standingResult struct {
		UserID string `db:"user_id"`
		Rank   int    `db:"rank"`
	}

	var standings []standingResult

	err = db.Select(&standings, `
		SELECT ss.user_id, ss.rank
		FROM season_standings ss
		WHERE ss.season_id = $1
		AND ss.week_id = (
			SELECT id FROM weeks WHERE season_id = $1 AND status = 'final' ORDER BY number DESC LIMIT 1
		)
		ORDER BY ss.rank ASC
	`, prevSeasonID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if len(standings) > 0 {
		prevYear := seasonYear - 1

		// winner - rank 1, handles ties
		for _, s := range standings {
			if s.Rank == 1 {
				badges[s.UserID] = append(badges[s.UserID], models.Badge{
					Type:  models.BadgePreviousSeasonWinner,
					Label: fmt.Sprintf("%d Champion", prevYear),
				})
			}
		}

		// loser - max rank, handles ties. skip if everyone is rank 1 (single player or full tie)
		maxRank := standings[len(standings)-1].Rank
		if maxRank > 1 {
			for _, s := range standings {
				if s.Rank == maxRank {
					badges[s.UserID] = append(badges[s.UserID], models.Badge{
						Type:  models.BadgePreviousSeasonLoser,
						Label: fmt.Sprintf("%d Loser", prevYear),
					})
				}
			}
		}
	}

	return badges, nil
}
