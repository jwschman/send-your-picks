package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"pawked.com/sendyourpicks/internal/api/middleware"
	"pawked.com/sendyourpicks/internal/id"
	"pawked.com/sendyourpicks/internal/models"
	"pawked.com/sendyourpicks/internal/service"
)

// creates a new season.  Takes the year number, number of weeks, and a bool for if it's postseason or not
func NewSeason(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		userID := middleware.GetUserID(c)

		// parse request body
		var req struct {
			Year           int      `json:"year"`
			NumberOfWeeks  int      `json:"number_of_weeks"`
			IsPostseason   bool     `json:"is_postseason"`
			ParticipantIDs []string `json:"participant_ids"` // optional: user IDs to add as initial participants
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// validate year... between 2000 and 2999 seems like a pretty good range
		if req.Year < 2000 || req.Year > 2999 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
			return
		}

		// default to 18 weeks (current NFL regular season) if not provided
		if req.NumberOfWeeks == 0 {
			req.NumberOfWeeks = 18
		}

		if req.NumberOfWeeks < 1 || req.NumberOfWeeks > 22 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid number of weeks"})
			return
		}

		// Start transaction
		tx, err := db.Beginx()
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer tx.Rollback()

		// Check if season already exists
		var seasonExists bool
		err = tx.QueryRow(
			`
			SELECT EXISTS (
				SELECT 1 FROM public.seasons WHERE year = $1 AND is_postseason = $2
			)
			`, req.Year, req.IsPostseason).Scan(&seasonExists)

		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if seasonExists {
			c.JSON(http.StatusConflict, gin.H{"error": "Season already exists"})
			return
		}

		// generate a new id
		seasonID, err := id.New()
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed generating ULID"})
			return
		}

		// Insert new season
		_, err = tx.Exec(
			`
			INSERT INTO public.seasons (id, year, number_of_weeks, is_postseason, created_by)
			VALUES ($1, $2, $3, $4, $5)
			`,
			seasonID,
			req.Year,
			req.NumberOfWeeks,
			req.IsPostseason,
			userID,
		)

		// error inserting into database
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create season"})
			return
		}

		// Insert initial participants if provided.
		// Done within the same transaction so season + participants are atomic.
		participantCount := 0
		for _, participantID := range req.ParticipantIDs {
			_, err := tx.Exec(`
				INSERT INTO public.season_participants (season_id, user_id)
				VALUES ($1, $2)
			`, seasonID, participantID)
			if err != nil {
				c.Error(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add participant"})
				return
			}
			participantCount++
		}

		if err := tx.Commit(); err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit season"})
			return
		}

		// everything worked
		c.JSON(http.StatusCreated, gin.H{
			"id":                seasonID,
			"year":              req.Year,
			"participant_count": participantCount, // how many participants were added
		})
	}

}

// Marks a season as active.  Fails if another season is already active
func ActivateSeason(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		seasonID := c.Param("season_id")

		// Start a transaction
		tx, err := db.Beginx()
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer tx.Rollback()

		// First, verify the season exists
		var seasonExists bool
		err = tx.Get(&seasonExists, `SELECT EXISTS(SELECT 1 FROM public.seasons WHERE id = $1)`, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		if !seasonExists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Season not found"})
			return
		}

		// Check if another season is already active
		var activeSeasonID string
		err = tx.Get(&activeSeasonID, `SELECT id FROM public.seasons WHERE is_active = true LIMIT 1`)
		// we can ignore ErrNoRows because that is what we actually want
		if err != nil && err != sql.ErrNoRows {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// If there's an active season and it's not the one we're trying to activate
		if err != sql.ErrNoRows && activeSeasonID != seasonID {
			c.JSON(http.StatusConflict, gin.H{
				"error":            "Another season is already active. Please deactivate it first.",
				"active_season_id": activeSeasonID,
			})
			return
		}

		// Activate the specified season
		_, err = tx.Exec(`UPDATE public.seasons SET is_active = true WHERE id = $1`, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate season"})
			return
		}

		// Commit the transaction
		err = tx.Commit()
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit changes"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"season_id": seasonID,
		})
	}
}

// DeactivateSeason marks a season as inactive.
// Returns 400 if the season is already inactive
func DeactivateSeason(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		seasonID := c.Param("season_id")

		// Check if season exists and get its current active status in one query
		var isActive bool
		err := db.Get(&isActive, `SELECT is_active FROM public.seasons WHERE id = $1`, seasonID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Season not found"})
				return
			}
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Don't deactivate an already inactive season
		if !isActive {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Season is already inactive"})
			return
		}

		// Deactivate the season.
		_, err = db.Exec(`UPDATE public.seasons SET is_active = false WHERE id = $1`, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate season"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"season_id": seasonID,
		})
	}
}

// UpdateSeasonWeeks updates the number of weeks in a season.
// Intended as a correction tool if the wrong value was entered at creation.
// Won't allow setting it lower than the number of weeks already created.
func UpdateSeasonWeeks(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		seasonID := c.Param("season_id")

		var req struct {
			NumberOfWeeks int `json:"number_of_weeks"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if req.NumberOfWeeks < 1 || req.NumberOfWeeks > 22 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Number of weeks must be between 1 and 22"})
			return
		}

		// Make sure the season exists
		seasonExists, err := service.SeasonExists(db, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		if !seasonExists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Season not found"})
			return
		}

		// Don't allow setting number_of_weeks below the number of weeks already created
		var existingWeekCount int
		err = db.Get(&existingWeekCount, `SELECT COUNT(*) FROM public.weeks WHERE season_id = $1`, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		if req.NumberOfWeeks < existingWeekCount {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Cannot set number of weeks to %d — %d weeks already exist", req.NumberOfWeeks, existingWeekCount),
			})
			return
		}

		_, err = db.Exec(`UPDATE public.seasons SET number_of_weeks = $1 WHERE id = $2`, req.NumberOfWeeks, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update season"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"season_id":       seasonID,
			"number_of_weeks": req.NumberOfWeeks,
		})
	}
}

// GetActiveSeason returns the currently active season
func GetActiveSeason(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var season models.Season

		query := `SELECT * FROM public.seasons WHERE is_active = true LIMIT 1`
		err := db.Get(&season, query)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "No active season"})
				return
			}
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// set context to include season id and year
		c.JSON(http.StatusOK, gin.H{
			"id":            season.ID,
			"year":          season.Year,
			"is_postseason": season.IsPostseason,
		})
	}
}

// returns all seasons
func GetAllSeasons(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		var seasons []models.Season

		// build the query first
		query := `SELECT id, year, is_active, number_of_weeks, is_postseason FROM public.seasons ORDER BY year`

		err := db.Select(&seasons, query) // select for multiple rows, Get for single row
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error",
			})
			return
		}

		// set context to include season id and year
		c.JSON(http.StatusOK, gin.H{
			"seasons": seasons,
		})

	}
}

// returns a single season by ID along with any weeks it contains
func GetSeason(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		seasonID := c.Param("season_id")

		type SeasonWithWeeks struct {
			ID            string        `json:"id" db:"id"`
			Year          int           `json:"year" db:"year"`
			IsActive      bool          `json:"is_active" db:"is_active"`
			NumberOfWeeks int           `json:"number_of_weeks" db:"number_of_weeks"`
			IsPostseason  bool          `json:"is_postseason" db:"is_postseason"`
			Weeks         []models.Week `json:"weeks"`
		}

		var season SeasonWithWeeks

		// Get season
		seasonQuery := `SELECT id, year, is_active, number_of_weeks, is_postseason FROM public.seasons WHERE id = $1`
		err := db.Get(&season, seasonQuery, seasonID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Season not found"})
				return
			}
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Get weeks
		weekQuery := `
            SELECT *
            FROM public.weeks 
            WHERE season_id = $1 
            ORDER BY number
        `
		var weeks []models.Week
		err = db.Select(&weeks, weekQuery, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Initialize empty array if nil
		if weeks == nil {
			weeks = []models.Week{}
		}
		season.Weeks = weeks

		c.JSON(http.StatusOK, gin.H{"season": season})
	}
}

// GetSeasonParticipants returns all users participating in a season.
// Joins with profiles to include display info (username, avatar, etc.)
func GetSeasonParticipants(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		seasonID := c.Param("season_id")

		// Use shared helper to verify season exists
		seasonExists, err := service.SeasonExists(db, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		if !seasonExists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Season not found"})
			return
		}

		// Database row struct (avatar_url is nullable)
		type participantRow struct {
			UserID    string    `db:"user_id"`
			Username  *string   `db:"username"`
			AvatarURL *string   `db:"avatar_url"`
			JoinedAt  time.Time `db:"joined_at"`
		}

		// Query participants joined with profile data for display.
		// Using LEFT JOIN in case a profile is somehow missing
		var rows []participantRow
		query := `
			SELECT
				sp.user_id,
				p.username,
				p.avatar_url,
				sp.joined_at
			FROM public.season_participants sp
			LEFT JOIN public.profiles p ON p.id = sp.user_id
			WHERE sp.season_id = $1
			ORDER BY sp.joined_at ASC
		`
		err = db.Select(&rows, query, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Response struct with full avatar URL
		type participantResponse struct {
			UserID    string    `json:"user_id"`
			Username  *string   `json:"username"`
			AvatarURL string    `json:"avatar_url"`
			JoinedAt  time.Time `json:"joined_at"`
		}

		// Transform rows to response, building full avatar URLs
		participants := make([]participantResponse, len(rows))
		for i, row := range rows {
			participants[i] = participantResponse{
				UserID:    row.UserID,
				Username:  row.Username,
				AvatarURL: buildAvatarURL(row.AvatarURL),
				JoinedAt:  row.JoinedAt,
			}
		}

		c.JSON(http.StatusOK, gin.H{"participants": participants})
	}
}

// AddSeasonParticipants adds one or more users to a season.
// Accepts an array of user IDs in the request body.
// Skips users who are already participants.
func AddSeasonParticipants(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		seasonID := c.Param("season_id")

		// Parse request body - expects array of user IDs
		var req struct {
			UserIDs []string `json:"user_ids"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Validate we have at least one user ID
		if len(req.UserIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_ids array is required and cannot be empty"})
			return
		}

		seasonExists, err := service.SeasonExists(db, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		if !seasonExists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Season not found"})
			return
		}

		// Start transaction
		tx, err := db.Beginx()
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer tx.Rollback()

		// Insert each participant.
		// ON CONFLICT DO NOTHING makes this idempotent
		// We track how many were actually added vs already existed.
		var addedCount int
		for _, userID := range req.UserIDs {
			result, err := tx.Exec(`
				INSERT INTO public.season_participants (season_id, user_id)
				VALUES ($1, $2)
				ON CONFLICT (season_id, user_id) DO NOTHING
			`, seasonID, userID)
			if err != nil {
				c.Error(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add participant"})
				return
			}

			// Check if a row was actually inserted (vs skipped due to conflict)
			rowsAffected, _ := result.RowsAffected()
			if rowsAffected > 0 {
				addedCount++
			}
		}

		// commit the transaction
		if err := tx.Commit(); err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit changes"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"added":           addedCount,                    // number of new participants added
			"already_existed": len(req.UserIDs) - addedCount, // number that were already in the season
			"total_requested": len(req.UserIDs),              // total user IDs in the request
		})
	}
}

// RemoveSeasonParticipant removes a single user from a season.
// Returns 404 if the user wasn't a participant.
// Note: This doesn't delete their picks or standings
func RemoveSeasonParticipant(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		seasonID := c.Param("season_id")
		userID := c.Param("user_id")

		// verify season exists
		seasonExists, err := service.SeasonExists(db, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		if !seasonExists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Season not found"})
			return
		}

		// Delete the participant record.
		result, err := db.Exec(`
			DELETE FROM public.season_participants
			WHERE season_id = $1 AND user_id = $2
		`, seasonID, userID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove participant"})
			return
		}

		// Check if a row was actually deleted
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "User is not a participant in this season"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"removed": true,
			"user_id": userID,
		})
	}
}
