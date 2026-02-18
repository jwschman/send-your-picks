package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"pawked.com/sendyourpicks/internal/logger"
	"pawked.com/sendyourpicks/internal/models"
	"pawked.com/sendyourpicks/internal/service"
)

// GetWeeks should be viewable by anyone, and will just return info about the weeks, including which is active
func GetWeeks(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		// assign seasonID
		seasonID := c.Param("season_id")
		if seasonID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing season ID"})
			return
		}

		// week summary shape
		type WeekSummary struct {
			ID     string `json:"id" db:"id"`
			Number int    `json:"number" db:"number"`
			Status string `json:"status" db:"status"`
		}

		var weeks []WeekSummary

		// build the query first
		query := `SELECT id, number, status FROM public.weeks WHERE season_id = $1 ORDER BY number`

		err := db.Select(&weeks, query, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database Error"})
			return
		}

		// set context to include season id and year
		c.JSON(http.StatusOK, gin.H{
			"season_id": seasonID,
			"weeks":     weeks,
		})
	}
}

// GetWeek returns metadata about a week and all games included in the week
func GetWeek(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		weekID := c.Param("week_id")
		if weekID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing week ID"})
			return
		}

		var week models.Week

		weekQuery := `
			SELECT
				w.id,
				w.season_id,
				w.number,
				w.created_at,
				w.updated_at,
				w.created_by,
				w.status,
				s.year,
				s.is_postseason
			FROM public.weeks w
			INNER JOIN public.seasons s ON w.season_id = s.id
			WHERE w.id = $1
		`

		err := db.Get(&week, weekQuery, weekID)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Week not found"})
			return
		}
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		var games []models.Game

		gamesQuery := `
			SELECT
				g.id,
				g.external_game_id,
				g.season_id,
				g.week_id,
				g.home_team_id,
				g.away_team_id,
				g.home_score,
				g.away_score,
				g.home_spread,
				g.kickoff_time,
				g.neutral_site,
				g.status,
				g.created_at,
				g.updated_at,
				g.created_by,

				ht.abbreviation AS home_team_abbr,
				at.abbreviation AS away_team_abbr,

				ht.name AS home_team_name,
				ht.city AS home_team_city,
				ht.logo_url AS home_team_logo_url,
				at.name AS away_team_name,
				at.city AS away_team_city,
				at.logo_url AS away_team_logo_url

			FROM public.games g
			INNER JOIN public.teams ht ON g.home_team_id = ht.id
			INNER JOIN public.teams at ON g.away_team_id = at.id
			WHERE g.week_id = $1
			ORDER BY g.kickoff_time
		`

		err = db.Select(&games, gamesQuery, weekID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// return empty slice if no games
		if games == nil {
			games = []models.Game{}
		}

		week.Games = games

		c.JSON(http.StatusOK, gin.H{"week": week})
	}
}

// just returns the status of a week
func GetWeekStatus(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		weekID := c.Param("week_id")
		if weekID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing week ID"})
			return
		}

		var week struct {
			ID     string `json:"id" db:"id"`
			Status string `json:"status" db:"status"`
		}

		query := `SELECT id, status FROM weeks WHERE id = $1`

		err := db.Get(&week, query, weekID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{
					"error":   "Week not found",
					"week_id": weekID,
				})
				return
			}
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database Error",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":     week.ID,
			"status": week.Status,
		})
	}
}

// manually updates the spreads for games in a week.  I think this belongs in week, not games, but I could see how it would go there also
func UpdateSpreads(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		weekID := c.Param("week_id")
		if weekID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing week ID"})
			return
		}

		type GameSpreadUpdate struct {
			GameID     string   `json:"game_id" binding:"required"`
			HomeSpread *float64 `json:"home_spread"`
		}

		type UpdateWeekRequest struct {
			Games []GameSpreadUpdate `json:"games" binding:"required"`
		}

		var req UpdateWeekRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		// Make sure the week exists
		weekExists, err := service.WeekExists(db, weekID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database Error"})
			return
		} else if !weekExists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Week not found", "week_id": weekID})
			return
		}

		// Actually fetch the week to check status
		var week models.Week
		err = db.Get(&week, `SELECT * FROM weeks WHERE id = $1`, weekID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Check if week is in games_imported or spreads_set status (can edit until activated)
		if week.Status != "games_imported" && week.Status != "spreads_set" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":          "Can only set spreads when week is in games_imported or spreads_set status",
				"current_status": week.Status,
			})
			return
		}

		// Validate all spreads before starting transaction
		for _, gameUpdate := range req.Games {
			if gameUpdate.HomeSpread != nil {
				spread := *gameUpdate.HomeSpread
				// Check if spread is a multiple of 0.5
				// Use modulo to check: valid spreads should have remainder of 0 when divided by 0.5
				remainder := spread - (float64(int(spread/0.5)) * 0.5)
				if remainder < -0.001 || remainder > 0.001 { // Small epsilon for floating point comparison
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "Invalid spread value. Spreads must be multiples of 0.5 (e.g., -3.5, 0, 1.5, 7.0)",
						"game_id": gameUpdate.GameID,
						"spread":  spread,
					})
					return
				}
			}
		}

		// Start transaction
		tx, err := db.Beginx()
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer tx.Rollback()

		// Update each game's spread
		for _, gameUpdate := range req.Games {
			query := `
                UPDATE games
                SET home_spread = $1, updated_at = NOW()
                WHERE id = $2 AND week_id = $3
            `
			result, err := tx.Exec(query, gameUpdate.HomeSpread, gameUpdate.GameID, weekID)
			if err != nil {
				c.Error(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update game"})
				return
			}

			rowsAffected, _ := result.RowsAffected()
			if rowsAffected == 0 {
				c.JSON(http.StatusNotFound, gin.H{"error": "Game not found or doesn't belong to this week"})
				return
			}
		}

		// Update week status to spreads_set
		_, err = tx.Exec(`UPDATE weeks SET status = 'spreads_set', updated_at = NOW() WHERE id = $1`, weekID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update week status"})
			return
		}

		if err := tx.Commit(); err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit changes"})
			return
		}

		// Fetch updated week to return
		var updatedWeek models.Week
		err = db.Get(&updatedWeek, `
            SELECT w.*, s.year, s.is_postseason
            FROM weeks w
            JOIN seasons s ON w.season_id = s.id
            WHERE w.id = $1
        `, weekID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		var games []models.Game
		err = db.Select(&games, `
            SELECT
                g.*,
                ht.name AS home_team_name,
                ht.city AS home_team_city,
                ht.logo_url AS home_team_logo_url,
                at.name AS away_team_name,
                at.city AS away_team_city,
                at.logo_url AS away_team_logo_url
            FROM games g
            JOIN teams ht ON g.home_team_id = ht.id
            JOIN teams at ON g.away_team_id = at.id
            WHERE g.week_id = $1
            ORDER BY g.kickoff_time
        `, weekID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if games == nil {
			games = []models.Game{}
		}
		updatedWeek.Games = games

		c.JSON(http.StatusOK, gin.H{"week": updatedWeek})
	}
}

// Just sets a week's status to active.  Currently used by the commissioner to release it to users
func ActivateWeek(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		weekID := c.Param("week_id")
		if weekID == "" {
			logger.Warn("activate week request missing week_id")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing week ID"})
			return
		}

		logger.Info("activate week request received", "week_id", weekID)

		// Start transaction to ensure consistent state during activation
		tx, err := db.Beginx()
		if err != nil {
			logger.Error("failed to begin transaction for activate week", "week_id", weekID, "error", err)
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer tx.Rollback()

		logger.Debug("transaction started for activate week", "week_id", weekID)

		// Check week exists and get current status (with lock)
		var week models.Week
		err = tx.Get(&week, `SELECT * FROM weeks WHERE id = $1 FOR UPDATE`, weekID)
		if err != nil {
			if err == sql.ErrNoRows {
				logger.Warn("activate week failed: week not found", "week_id", weekID)
				c.JSON(http.StatusNotFound, gin.H{"error": "Week not found", "week_id": weekID})
				return
			}
			logger.Error("failed to fetch week for activation", "week_id", weekID, "error", err)
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		logger.Debug(
			"week loaded for activation",
			"week_id", weekID,
			"current_status", week.Status,
		)

		// Check if week is in spreads_set status
		if week.Status != "spreads_set" {
			logger.Warn(
				"activate week rejected due to invalid status",
				"week_id", weekID,
				"current_status", week.Status,
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":          "Week is not in spreads_set status",
				"current_status": week.Status,
			})
			return
		}

		// Make sure all games have a spread
		var gamesWithoutSpreads int
		err = tx.Get(&gamesWithoutSpreads, `
            SELECT COUNT(*)
            FROM games
            WHERE week_id = $1 AND home_spread IS NULL
        `, weekID)
		if err != nil {
			logger.Error(
				"failed to validate game spreads before activating week",
				"week_id", weekID,
				"error", err,
			)
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if gamesWithoutSpreads > 0 {
			logger.Warn(
				"activate week blocked due to missing spreads",
				"week_id", weekID,
				"games_missing_spreads", gamesWithoutSpreads,
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":                 "Cannot activate week: some games are missing spreads",
				"games_missing_spreads": gamesWithoutSpreads,
			})
			return
		}

		// Update week status to active
		_, err = tx.Exec(`
            UPDATE weeks
            SET status = 'active', activated_at = NOW(), updated_at = NOW()
            WHERE id = $1
        `, weekID)
		if err != nil {
			logger.Error(
				"failed to update week status to active",
				"week_id", weekID,
				"error", err,
			)
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate week"})
			return
		}

		logger.Info("week status updated to active", "week_id", weekID)

		if err := tx.Commit(); err != nil {
			logger.Error(
				"failed to commit activate week transaction",
				"week_id", weekID,
				"error", err,
			)
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit activation"})
			return
		}

		logger.Info("week successfully activated", "week_id", weekID)

		c.JSON(http.StatusOK, gin.H{
			"week_id": weekID,
			"status":  "active",
		})
	}
}

// GetActiveWeek returns the currently active week
func GetActiveWeek(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		seasonID := c.Param("season_id")
		if seasonID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing season ID"})
			return
		}

		var activeWeek struct {
			Id     string `json:"id" db:"id"`
			Number int    `json:"number" db:"number"`
		}

		query := `SELECT id, number
					FROM public.weeks
					WHERE season_id = $1
						AND status = 'active'
					ORDER BY number DESC
					LIMIT 1;
				` // do i need the LIMIT 1?  And the order?  Probably... just in case there is more than one active week (there shouldnt be)

		err := db.Get(&activeWeek, query, seasonID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "No active week"})
				return
			}
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// set context to include week id
		c.JSON(http.StatusOK, gin.H{
			"id":     activeWeek.Id,
			"number": activeWeek.Number,
		})
	}
}
