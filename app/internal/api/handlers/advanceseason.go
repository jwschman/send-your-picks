package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"pawked.com/sendyourpicks/internal/api/middleware"
	"pawked.com/sendyourpicks/internal/logger"
	"pawked.com/sendyourpicks/internal/models"
	"pawked.com/sendyourpicks/internal/service"
)

// Advances the season's state
// To be run as a cron job but can also be started manually now for testing
func AdvanceSeason(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// we're still using userID to set who modified it, but it will all be automatic later
		userID := middleware.GetUserID(c)

		seasonID := c.Param("season_id")

		// first we check if there are any nonfinal weeks
		var week models.Week
		err := db.Get(&week, `
			SELECT *
			FROM weeks
			WHERE season_id = $1
				AND status != 'final'
			LIMIT 1
		`, seasonID)

		// an error of no rows would mean there are no final weeks
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// No active week so we create one
				logger.Info("No active week.  Creating next week")

				newWeek, err := service.CreateNextWeekForSeason(db, seasonID, userID)
				if errors.Is(err, service.ErrSeasonComplete) {
					c.JSON(http.StatusOK, gin.H{
						"action": "season_complete",
					})
					return
				}
				if err != nil {
					c.Error(err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create next week"})
					return
				}

				// set week to the newly created week
				week = *newWeek

			} else {
				// actual error loading the week
				c.Error(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load week"})
				return
			}
		}

		// A non-final week exists so we advance it
		logger.Debug("A non-final week exists.  Calling AdvanceWeekState", "week_number", week.Number)
		if err := service.AdvanceWeekState(c.Request.Context(), db, &week, userID); err != nil {
			// Handle manual action required - not really an error, just needs commissioner input
			if errors.Is(err, service.ErrManualActionRequired) {
				if reloadErr := db.Get(&week, `SELECT * FROM weeks WHERE id = $1`, week.ID); reloadErr != nil {
					c.Error(reloadErr)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload week"})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"action": "manual_action_required",
					"week":   week.ID,
					"status": week.Status,
				})
				return
			}

			// Handle waiting for games to finish
			if errors.Is(err, service.ErrWaitingForGames) {
				if reloadErr := db.Get(&week, `SELECT * FROM weeks WHERE id = $1`, week.ID); reloadErr != nil {
					c.Error(reloadErr)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload week"})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"action": "waiting_for_games",
					"week":   week.ID,
					"status": week.Status,
				})
				return
			}

			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to advance week"})
			return
		}

		// Reload week to get updated status
		if err := db.Get(&week, `SELECT * FROM weeks WHERE id = $1`, week.ID); err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload week"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"action": "week_advanced",
			"week":   week.ID,
			"status": week.Status,
		})
	}
}
