package handlers

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"pawked.com/sendyourpicks/internal/logger"
	"pawked.com/sendyourpicks/internal/models"
	"pawked.com/sendyourpicks/internal/external"
	"pawked.com/sendyourpicks/internal/service"
)

// timesAreClose checks if two times are within the tolerance window
// This handles timezone differences and slight schedule adjustments
func timesAreClose(t1, t2 time.Time, tolerance time.Duration) bool {
	diff := math.Abs(float64(t1.Sub(t2)))
	return diff <= float64(tolerance)
}

// AutoImportSpreads automatically fetches and sets spreads from the Odds API
// I wish I could test this now, but it did work for the super bowl
func AutoImportSpreads(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		weekID := c.Param("week_id")
		if weekID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing week ID"})
			return
		}

		// specify preferred bookmaker in query param.  Maybe this shoul be set in global settings or even ENV?
		bookmaker := c.DefaultQuery("bookmaker", "draftkings")

		logger.Info(
			"auto-importing spreads from odds API",
			"week_id", weekID,
			"bookmaker", bookmaker,
		)

		// Check week exists and get status
		weekStatus, err := service.GetWeekStatus(db, weekID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Check if week is in games_imported or spreads_set status
		if weekStatus != "games_imported" && weekStatus != "spreads_set" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":          "Can only import spreads when week is in games_imported or spreads_set status",
				"current_status": weekStatus,
			})
			return
		}

		// Get all games for this week.  I should really make a getGames function in my services
		var games []models.Game
		err = db.Select(&games, `
			SELECT
				g.id,
				g.home_team_id,
				g.away_team_id,
				g.home_spread,
				g.kickoff_time,
				ht.abbreviation as home_team_abbr,
				at.abbreviation as away_team_abbr,
				ht.name as home_team_name,
				at.name as away_team_name
			FROM games g
			JOIN teams ht ON g.home_team_id = ht.id
			JOIN teams at ON g.away_team_id = at.id
			WHERE g.week_id = $1
		`, weekID)
		if err != nil {
			logger.Error(
				"failed to fetch games for week",
				"week_id", weekID,
				"error", err,
			)
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if len(games) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No games found for this week"})
			return
		}

		logger.Info(
			"fetched games for spread import",
			"week_id", weekID,
			"game_count", len(games),
		)

		// Create Odds API client
		oddsClient, err := external.NewOddsClient()
		if err != nil {
			logger.Error(
				"failed to create odds API client",
				"error", err,
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to initialize Odds API client",
				"details": err.Error(),
			})
			return
		}

		// Fetch spreads from Odds API
		spreads, err := oddsClient.FetchSpreads(c.Request.Context(), bookmaker)
		if err != nil {
			logger.Error(
				"failed to fetch spreads from odds API",
				"error", err,
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch spreads from Odds API",
				"details": err.Error(),
			})
			return
		}

		logger.Info(
			"fetched spreads from odds API",
			"spread_count", len(spreads),
		)

		// Match games with spreads using team matchup AND kickoff time
		// Tolerance of 6 hours handles timezone differences and slight schedule adjustments
		const kickoffTolerance = 6 * time.Hour

		type GameUpdate struct {
			GameID     string
			HomeSpread float64
			Matched    bool
		}

		updates := make([]GameUpdate, 0, len(games))
		matchedCount := 0
		unmatchedGames := []string{}

		// range over all the games and attach the spreads
		for _, game := range games {
			var matchedSpread *external.SpreadInfo

			// Look for matching spread by team matchup AND kickoff time
			for i := range spreads {
				spread := &spreads[i]

				// Check if teams match
				teamsMatch := (spread.HomeTeamAbbr == game.HomeTeamAbbr &&
					spread.AwayTeamAbbr == game.AwayTeamAbbr)

				if !teamsMatch {
					continue
				}

				// Check if kickoff times are close
				if timesAreClose(spread.CommenceTime, game.KickoffTime, kickoffTolerance) {
					matchedSpread = spread
					logger.Info(
						"matched game with spread",
						"game_id", game.ID,
						"matchup", fmt.Sprintf("%s @ %s", game.AwayTeamAbbr, game.HomeTeamAbbr),
						"spread", spread.HomeSpread,
						"db_kickoff", game.KickoffTime.Format(time.RFC3339),
						"odds_kickoff", spread.CommenceTime.Format(time.RFC3339),
					)
					break
				} else {
					logger.Warn(
						"teams match but kickoff times too different",
						"game_id", game.ID,
						"matchup", fmt.Sprintf("%s @ %s", game.AwayTeamAbbr, game.HomeTeamAbbr),
						"db_kickoff", game.KickoffTime.Format(time.RFC3339),
						"odds_kickoff", spread.CommenceTime.Format(time.RFC3339),
						"time_diff_hours", math.Abs(float64(spread.CommenceTime.Sub(game.KickoffTime).Hours())),
					)
				}
			}

			if matchedSpread != nil {
				updates = append(updates, GameUpdate{
					GameID:     game.ID,
					HomeSpread: matchedSpread.HomeSpread,
					Matched:    true,
				})
				matchedCount++
			} else {
				unmatchedGames = append(unmatchedGames, fmt.Sprintf("%s @ %s (kickoff: %s)",
					game.AwayTeamAbbr, game.HomeTeamAbbr, game.KickoffTime.Format("Jan 2 3:04PM MST")))
				logger.Warn(
					"no spread found for game",
					"game_id", game.ID,
					"matchup", fmt.Sprintf("%s @ %s", game.AwayTeamAbbr, game.HomeTeamAbbr),
					"kickoff_time", game.KickoffTime.Format(time.RFC3339),
				)
			}
		}

		// If no games matched, return error
		if matchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "No spreads found for any games in this week",
				"details": "This might mean the Odds API doesn't have data for these games yet",
			})
			return
		}

		// Start transaction to update spreads
		tx, err := db.Beginx()
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer tx.Rollback()

		// Update each matched game
		for _, update := range updates {
			if !update.Matched {
				continue
			}

			query := `
				UPDATE games
				SET home_spread = $1, updated_at = NOW()
				WHERE id = $2 AND week_id = $3
			`
			result, err := tx.Exec(query, update.HomeSpread, update.GameID, weekID)
			if err != nil {
				logger.Error(
					"failed to update game spread",
					"game_id", update.GameID,
					"error", err,
				)
				c.Error(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update game"})
				return
			}

			rowsAffected, _ := result.RowsAffected()
			if rowsAffected == 0 {
				logger.Warn(
					"game update affected 0 rows",
					"game_id", update.GameID,
				)
			}
		}

		// Update week status to spreads_set
		_, err = tx.Exec(`UPDATE weeks SET status = 'spreads_set', updated_at = NOW() WHERE id = $1`, weekID)
		if err != nil {
			logger.Error(
				"failed to update week status",
				"week_id", weekID,
				"error", err,
			)
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update week status"})
			return
		}

		if err := tx.Commit(); err != nil {
			logger.Error(
				"failed to commit spread updates",
				"week_id", weekID,
				"error", err,
			)
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit changes"})
			return
		}

		logger.Info(
			"successfully imported spreads",
			"week_id", weekID,
			"games_updated", matchedCount,
			"games_total", len(games),
		)

		// set up the response here
		response := gin.H{
			"games_updated": matchedCount,
			"games_total":   len(games),
			"bookmaker":     bookmaker,
			"week_status":   "spreads_set",
		}

		if len(unmatchedGames) > 0 {
			response["unmatched_games"] = unmatchedGames
			response["warning"] = fmt.Sprintf("%d games could not be matched with Odds API data", len(unmatchedGames))
		}

		c.JSON(http.StatusOK, response)
	}
}
