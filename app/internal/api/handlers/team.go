package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"pawked.com/sendyourpicks/internal/models"
)

// GetAllTeams returns all active teams in the database
func GetAllTeams(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var teams []models.Team

		query := `SELECT * FROM public.teams WHERE is_active = true ORDER BY city, name`
		err := db.Select(&teams, query)

		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if teams == nil {
			teams = []models.Team{}
		}

		c.JSON(http.StatusOK, gin.H{"teams": teams})
	}
}
