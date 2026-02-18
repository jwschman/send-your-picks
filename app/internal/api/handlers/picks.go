package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"pawked.com/sendyourpicks/internal/api/middleware"
	"pawked.com/sendyourpicks/internal/id"
	"pawked.com/sendyourpicks/internal/models"
	"pawked.com/sendyourpicks/internal/service"
	"pawked.com/sendyourpicks/internal/settings"
)

// this wound up hopefully being the most difficult thing in the whole app.  Wow.  TONS of checking.

// creates picks from request
func SubmitPicks(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		userID := middleware.GetUserID(c)

		weekID := c.Param("week_id")
		if weekID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing week ID"})
			return
		}

		// Check if user is a participant in this season.
		// Non-participants can view but not submit picks.
		isParticipant, err := service.IsUserSeasonParticipant(db, weekID, userID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		if !isParticipant {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not a participant in this season"})
			return
		}

		// get global settings (lockout times)
		settings, err := settings.Get(db)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load global settings"})
			return
		}

		// parse request
		type PickSubmission struct {
			GameID         string  `json:"game_id" binding:"required"`
			SelectedTeamID *string `json:"selected_team_id" binding:"required"`
		}

		type SubmitPicksRequest struct {
			Picks []PickSubmission `json:"picks" binding:"required"`
		}

		var req SubmitPicksRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// if there are no picks
		if len(req.Picks) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No picks provided"})
			return
		}

		// make sure there are no double picks
		seenGameIDs := make(map[string]bool)
		for _, pick := range req.Picks {
			if seenGameIDs[pick.GameID] {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Duplicate pick for game",
					"game_id": pick.GameID,
				})
				return
			}
			// Mark this game ID as seen
			seenGameIDs[pick.GameID] = true
		}

		// start transaction
		tx, err := db.Beginx()
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer tx.Rollback()

		// Load all games for the week
		type GameInfo struct {
			ID          string    `db:"id"`
			KickoffTime time.Time `db:"kickoff_time"`
			HomeTeamID  string    `db:"home_team_id"`
			AwayTeamID  string    `db:"away_team_id"`
		}

		var databaseGames []GameInfo

		query := `
			SELECT id, kickoff_time, home_team_id, away_team_id
			FROM games
			WHERE week_id = $1
			FOR SHARE
		`

		if err := tx.Select(&databaseGames, query, weekID); err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// convert databaseGames slice to map of IDs for faster lookup
		gameMap := make(map[string]GameInfo)
		for _, game := range databaseGames {
			gameMap[game.ID] = game
		}
		// Games are now loaded and we can begin validation

		// Validate picks (advisory only)
		for _, pick := range req.Picks {
			game, exists := gameMap[pick.GameID]
			if !exists {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid game",
					"game_id": pick.GameID,
				})
				return
			}

			// Check if pick is still allowed (skip if testing mode (-1))
			if settings.PickCutoffMinutes != -1 {
				// time math by claude... this stuff is hard
				// all times are UTC
				cutoff := game.KickoffTime.Add(-time.Duration(settings.PickCutoffMinutes) * time.Minute)
				if time.Now().UTC().After(cutoff) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "This game's pick window has closed",
						"game_id": pick.GameID,
					})
					return
				}
			}

			// make sure the team is a valid choice - allow nil choice
			if pick.SelectedTeamID != nil {
				if *pick.SelectedTeamID != game.HomeTeamID && *pick.SelectedTeamID != game.AwayTeamID {
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "Selected team does not belong to this game",
						"game_id": pick.GameID,
					})
					return
				}
			}
		}

		// Insert / update picks
		for _, pick := range req.Picks {

			// we're generating a pickID every time.  It will be unnecessary if the pick already exists, but whatever
			pickID, err := id.New()
			if err != nil {
				c.Error(err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Error generating pick ID",
					"game_id": pick.GameID,
				})
				return
			}

			// once again query by claude
			// the ON CONFLICT user_id game_id is what makes sure that there's only user pick per game, and where the id will get skipped
			// because DO UPDATE SET will still actually update the team id of the pick.  nothing else needs to be udpdated
			query := `
				INSERT INTO picks (id, user_id, game_id, week_id, selected_team_id)
				SELECT $1, $2, $3, $4, $5
				FROM games g
				WHERE g.id = $3
				AND (
					$6 = true
					OR now() < g.kickoff_time
				)
				ON CONFLICT (user_id, game_id)
				DO UPDATE
				SET selected_team_id = EXCLUDED.selected_team_id
				WHERE
				picks.user_locked_at IS NULL
				AND (
					$6 = true
					OR now() < (
					SELECT kickoff_time
					FROM games
					WHERE id = picks.game_id
					)
				);
			`

			result, err := tx.Exec(query, pickID, userID, pick.GameID, weekID, pick.SelectedTeamID, settings.AllowPicksAfterKickoff)
			if err != nil {
				c.Error(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save pick"})
				return
			}

			rows, err := result.RowsAffected()
			if err != nil {
				c.Error(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
				return
			}

			if rows == 0 {
				c.JSON(http.StatusConflict, gin.H{
					"error":   "Pick is locked or game has started",
					"game_id": pick.GameID,
				})
				return
			}
		}

		if err := tx.Commit(); err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit picks"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"picks_count": len(req.Picks),
		})
	}
}

// locks a users picks
func LockWeekPicks(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		userID := middleware.GetUserID(c)

		weekID := c.Param("week_id")
		if weekID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing week ID"})
			return
		}

		// Check if user is a participant in this season.
		// Non-participants cannot lock picks (they shouldn't have any anyway).
		isParticipant, err := service.IsUserSeasonParticipant(db, weekID, userID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		if !isParticipant {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not a participant in this season"})
			return
		}

		// set up the transaction
		tx, err := db.Beginx()
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer tx.Rollback()

		// struct to hold the status of picks
		var pickLockStatus struct {
			Total      int `db:"total"`      // total number of picks
			Unlocked   int `db:"unlocked"`   // total number of unlocked picks (should always equal total)
			Incomplete int `db:"incomplete"` // total number of picks that don't have a team selected
		}

		// check pick status
		query := `
			SELECT
				COUNT(*) AS total,
				COUNT(*) FILTER (WHERE user_locked_at IS NULL) AS unlocked,
				COUNT(*) FILTER (
					WHERE user_locked_at IS NULL
					AND selected_team_id IS NULL
				) AS incomplete
			FROM picks
			WHERE user_id = $1
			AND week_id = $2
			`

		err = tx.Get(&pickLockStatus, query, userID, weekID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate picks"})
			return
		}

		if pickLockStatus.Total == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No picks to lock"})
			return
		}

		if pickLockStatus.Unlocked != pickLockStatus.Total {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Picks already locked"})
			return
		}

		if pickLockStatus.Incomplete > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "All picks must be selected before locking"})
			return
		}

		// Actually lock the picks by setting user_locked_at
		_, err = tx.Exec(`
			UPDATE picks
			SET user_locked_at = NOW()
			WHERE user_id = $1
			AND week_id = $2
			AND user_locked_at IS NULL
		`, userID, weekID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to lock picks"})
			return
		}

		// commit the transaction
		if err := tx.Commit(); err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit lock"})
			return
		}

		// response
		c.JSON(http.StatusOK, gin.H{
			"locked_picks": pickLockStatus.Total,
		})
	}
}

// return a users already created picks for a week
func GetMyPicks(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		userID := middleware.GetUserID(c)

		weekID := c.Param("week_id")
		if weekID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing week ID"})
			return
		}

		var userPicks []models.Pick

		query := `SELECT * from public.picks WHERE week_id = $1 AND user_id = $2`

		err := db.Select(&userPicks, query, weekID, userID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database Error"})
			return
		}
		// Initialize empty array if nil
		if userPicks == nil {
			userPicks = []models.Pick{}
		}

		c.JSON(http.StatusOK, gin.H{"picks": userPicks})
	}
}

// returns a summary of the logged in user's picks
func GetMyWeekPickSummary(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		weekID := c.Param("week_id")
		if weekID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing week ID"})
			return
		}

		query := `
			SELECT
				$1 AS week_id,
				COUNT(g.id) AS total_games,
				COUNT(p.id) FILTER (WHERE p.selected_team_id IS NOT NULL) AS picks_completed,
				COUNT(p.id) FILTER (WHERE p.selected_team_id IS NOT NULL) = COUNT(g.id) AS all_picks_completed,
				COALESCE(
					BOOL_AND(p.user_locked_at IS NOT NULL)
						FILTER (WHERE p.selected_team_id IS NOT NULL),
					false
				) AS all_picks_locked
			FROM public.games g
			LEFT JOIN public.picks p
				ON p.game_id = g.id
				AND p.user_id = $2
			WHERE g.week_id = $1;
		`

		var summary models.PickSummary
		err := db.Get(&summary, query, weekID, userID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database Error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"summary": summary,
		})
	}
}

// returns a summary of user picks for a week
func GetWeekPickSummary(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		weekID := c.Param("week_id")
		if weekID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing week ID"})
			return
		}

		// Get total games for the week
		var totalGames int
		err := db.Get(&totalGames, `SELECT COUNT(*) FROM public.games WHERE week_id = $1`, weekID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database Error"})
			return
		}

		type UserPickSummary struct {
			UserID         string `json:"user_id" db:"user_id"`
			Username       string `json:"username" db:"username"`
			PicksSubmitted int    `json:"picks_submitted" db:"picks_submitted"`
		}

		var userPickSummary []UserPickSummary

		// query by Claude
		query := `
            SELECT
                p.id as user_id,
                p.username,
                COALESCE(COUNT(pk.id) FILTER (WHERE pk.selected_team_id IS NOT NULL), 0) as picks_submitted
            FROM public.profiles p
            LEFT JOIN public.picks pk ON pk.user_id = p.id AND pk.week_id = $1
            GROUP BY p.id, p.username
            ORDER BY picks_submitted DESC, p.username ASC
        `

		err = db.Select(&userPickSummary, query, weekID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database Error"})
			return
		}

		// Initialize empty array if nil
		if userPickSummary == nil {
			userPickSummary = []UserPickSummary{}
		}

		c.JSON(http.StatusOK, gin.H{
			"week_id":     weekID,
			"total_games": totalGames,
			"users":       userPickSummary,
		})
	}
}

// returns all users with their locked picks for a week
func GetWeekLockedPicks(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		weekID := c.Param("week_id")
		if weekID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing week ID"})
			return
		}

		// Pick detail for a specific game
		type PickDetail struct {
			GameID         string `json:"game_id" db:"game_id"`
			SelectedTeamID string `json:"selected_team_id" db:"selected_team_id"`
			IsCorrect      *bool  `json:"is_correct" db:"is_correct"`
		}

		// User with their picks and avatar
		type UserWithPicks struct {
			UserID    string       `json:"user_id"`
			Username  string       `json:"username"`
			AvatarURL *string      `json:"avatar_url"`
			Picks     []PickDetail `json:"picks"`
		}

		// Profile from db
		type Profile struct {
			ID        string  `db:"id"`
			Username  string  `db:"username"`
			AvatarURL *string `db:"avatar_url"`
		}

		// Locked pick from db
		type LockedPick struct {
			UserID         string  `db:"user_id"`
			GameID         *string `db:"game_id"`
			SelectedTeamID *string `db:"selected_team_id"`
			IsCorrect      *bool   `db:"is_correct"`
		}

		// Query 1: Get all users
		var profiles []Profile
		profilesQuery := `
			SELECT id, username, avatar_url
			FROM public.profiles
			ORDER BY username
		`
		err := db.Select(&profiles, profilesQuery)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database Error"})
			return
		}

		// Query 2: Get picks that are "visible" - either explicitly locked OR game has started
		var lockedPicks []LockedPick
		picksQuery := `
			SELECT
				p.user_id,
				p.game_id,
				p.selected_team_id,
				p.is_correct
			FROM public.picks p
			JOIN public.games g ON g.id = p.game_id
			WHERE p.week_id = $1
				AND p.selected_team_id IS NOT NULL
				AND (
					p.user_locked_at IS NOT NULL
					OR g.kickoff_time <= NOW()
				)
			ORDER BY p.user_id, p.game_id
		`
		err = db.Select(&lockedPicks, picksQuery, weekID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database Error"})
			return
		}

		// Build a map of user_id -> picks
		picksByUser := make(map[string][]PickDetail)
		for _, pick := range lockedPicks {
			if pick.GameID != nil && pick.SelectedTeamID != nil {
				picksByUser[pick.UserID] = append(picksByUser[pick.UserID], PickDetail{
					GameID:         *pick.GameID,
					SelectedTeamID: *pick.SelectedTeamID,
					IsCorrect:      pick.IsCorrect,
				})
			}
		}

		// Combine profiles with their picks
		users := make([]UserWithPicks, 0, len(profiles))
		for _, profile := range profiles {
			picks := picksByUser[profile.ID]
			if picks == nil {
				picks = []PickDetail{}
			}
			users = append(users, UserWithPicks{
				UserID:    profile.ID,
				Username:  profile.Username,
				AvatarURL: profile.AvatarURL,
				Picks:     picks,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"users": users,
		})
	}
}
