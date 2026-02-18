package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"pawked.com/sendyourpicks/internal/models"
)

// returns all of the current global settings
func GetGlobalSettings(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		var settings models.Settings

		// build the query first
		query := `SELECT
				id,
				pick_cutoff_minutes,
				allow_pick_edits,
				points_per_correct_pick,
				competition_timezone,
				allow_commissioner_overrides,
				allow_picks_after_kickoff,
				debug_mode,
				created_at,
				updated_at
			FROM public.settings
		`

		err := db.Get(&settings, query) // select for multiple rows, Get for single row
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database Error",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"settings": settings,
		})

	}
}

// updates the global settings db
func UpdateGlobalSettings(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		var req struct {
			Id                         string `json:"id"`
			PickCutoffMinutes          int    `json:"pick_cutoff_minutes"`
			AllowPickEdits             bool   `json:"allow_pick_edits"`
			PointsPerCorrectPick       int    `json:"points_per_correct_pick"`
			AllowCommissionerOverrides bool   `json:"allow_commissioner_overrides"`
			CompetitionTimezone        string `json:"competition_timezone"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if req.CompetitionTimezone == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "competition_timezone is required",
			})
			return
		}

		if _, err := time.LoadLocation(req.CompetitionTimezone); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid competition_timezone",
			})
			return
		}

		// set up transaction
		tx, err := db.Beginx()
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer tx.Rollback()

		var settings models.Settings

		// build the query first
		query := `
            UPDATE public.settings
            SET pick_cutoff_minutes = $2, allow_pick_edits = $3, points_per_correct_pick = $4, allow_commissioner_overrides = $5, competition_timezone = $6
            WHERE id = $1
            RETURNING *
		`

		err = tx.Get(&settings, query, req.Id, req.PickCutoffMinutes, req.AllowPickEdits, req.PointsPerCorrectPick, req.AllowCommissionerOverrides, req.CompetitionTimezone)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database Error",
			})
			return
		}

		if err := tx.Commit(); err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit update"})
			return
		}

		// return the updated settings
		c.JSON(http.StatusOK, gin.H{
			"settings": settings,
		})
	}
}
