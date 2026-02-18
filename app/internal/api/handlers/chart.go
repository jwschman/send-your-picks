package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// Returns the season's history to be graphed with chart.js
func GetSeasonHistory(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get season_id from URL params
		seasonID := c.Param("season_id")

		// Query standings history, filtered to only include season participants.
		query := `
			SELECT
				ss.user_id,
				p.username,
				ss.week_id,
				ss.points,
				ss.rank,
				ss.computed_at
			FROM public.season_standings ss
			JOIN public.profiles p ON ss.user_id = p.id
			JOIN public.season_participants sp ON sp.season_id = $1 AND sp.user_id = ss.user_id
			WHERE ss.season_id = $1
			ORDER BY ss.computed_at ASC, ss.rank ASC
		`

		// Define a struct for each row
		type StandingsHistoryEntry struct {
			UserID     string    `db:"user_id" json:"user_id"`
			Username   string    `db:"username" json:"username"`
			WeekID     string    `db:"week_id" json:"week_id"`
			Points     int       `db:"points" json:"points"`
			Rank       int       `db:"rank" json:"rank"`
			ComputedAt time.Time `db:"computed_at" json:"computed_at"`
		}

		var entries []StandingsHistoryEntry
		err := db.Select(&entries, query, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch standings history"})
			return
		}

		if entries == nil {
			entries = []StandingsHistoryEntry{}
		}

		c.JSON(http.StatusOK, gin.H{
			"history": entries,
		})
	}
}
