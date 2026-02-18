package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"pawked.com/sendyourpicks/internal/models"
	"pawked.com/sendyourpicks/internal/service"
)

// GetBadges returns badges for all users based on the active season.
func GetBadges(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// find the active season
		var seasonID string
		err := db.Get(&seasonID, `SELECT id FROM public.seasons WHERE is_active = true LIMIT 1`)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusOK, gin.H{"badges": map[string][]models.Badge{}})
				return
			}
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		badges, err := service.GetBadges(db, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load badges"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"badges": badges})
	}
}
