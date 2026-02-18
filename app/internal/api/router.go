package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"pawked.com/sendyourpicks/internal/api/handlers"
	"pawked.com/sendyourpicks/internal/api/middleware"
)

func RegisterRoutes(r *gin.Engine, db *sqlx.DB, authConfig *middleware.AuthConfig) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// All /api routes require authentication
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(authConfig))

	// Regular user routes (any authenticated user)

	// settings
	api.GET("/settings", handlers.GetGlobalSettings(db)) // get the global settings.  add season and week later if necessary

	// about the user
	api.GET("/whoami", handlers.WhoAmI(db))         // whoami data
	api.GET("/account", handlers.GetMyAccount(db))  // get info about my account
	api.PUT("/account", handlers.UpdateAccount(db)) // update my account

	// public user accounts
	api.GET("/users", handlers.GetAllUsers(db))      // list of all users
	api.GET("/users/:user_id", handlers.GetUser(db)) // public info about a single user
	api.GET("/badges", handlers.GetBadges(db))       // badges for all users (based on active season)

	// teams related
	api.GET("/teams", handlers.GetAllTeams(db)) // lists all active teams

	// Get info about seasons and weeks
	api.GET("/seasons", handlers.GetAllSeasons(db))                                 // lists metadata for all seasons
	api.GET("/seasons/active", handlers.GetActiveSeason(db))                        // see active season
	api.GET("/seasons/:season_id", handlers.GetSeason(db))                          // returns metadata and games for a single season
	api.GET("/seasons/:season_id/weeks/active", handlers.GetActiveWeek(db))         // get an active week if it exists
	api.GET("/seasons/:season_id/participants", handlers.GetSeasonParticipants(db)) // list users participating in this season
	api.GET("/weeks/:week_id", handlers.GetWeek(db))                                // get all games for a week

	// Picks related
	api.PUT("/weeks/:week_id/picks", handlers.SubmitPicks(db))                  // make and edit picks for logged in user
	api.GET("/weeks/:week_id/picks", handlers.GetMyPicks(db))                   // get my picks for the week
	api.GET("/weeks/:week_id/picks/summary", handlers.GetMyWeekPickSummary(db)) // returns a summary of my picks for a week
	api.POST("/weeks/:week_id/picks/lock", handlers.LockWeekPicks(db))          // locks all a user's picks for the week
	api.GET("/weeks/:week_id/picks/locked", handlers.GetWeekLockedPicks(db))    // returns all locked picks for all users for the week

	// Points and standings related
	api.GET("/weeks/:week_id/results", handlers.GetWeekResults(db))                       // get the results for a week for all users
	api.GET("/weeks/:week_id/standings", handlers.GetSeasonStandings(db))                 // gets the season standings after a given week
	api.GET("/seasons/:season_id/points", handlers.GetMySeasonPoints(db))                 // returns my per week points and standings
	api.GET("/seasons/:season_id/standings", handlers.GetCurrentSeasonStandings(db))      // Current: latest standings for the season
	api.GET("/seasons/:season_id/standings/me", handlers.GetMyCurrentSeasonStandings(db)) // gets logged in users current standings for the season
	api.GET("/seasons/:season_id/week-winners", handlers.GetWeekWinners(db))              // who won each week (with ties)
	api.GET("/seasons/:season_id/win-counts", handlers.GetUserWinCounts(db))              // user win/tie counts

	// Chart related
	api.GET("/seasons/:season_id/standings/history", handlers.GetSeasonHistory(db)) // returns the point and ranking history of the season for graphing

	// Commissioner routes (commissioner or admin)
	commissioner := api.Group("/commissioner")
	commissioner.Use(middleware.RequireRole("commissioner", "admin"))
	{
		// Season Management
		commissioner.POST("/seasons", handlers.NewSeason(db))                                 // create a new season
		commissioner.POST("/seasons/:season_id/advance", handlers.AdvanceSeason(db))          // advance the state of the season (state machine)
		commissioner.PATCH("/seasons/:season_id/activate", handlers.ActivateSeason(db))       // sets the active season
		commissioner.PATCH("/seasons/:season_id/deactivate", handlers.DeactivateSeason(db))   // deactivates the active season
		commissioner.PATCH("/seasons/:season_id/weeks-count", handlers.UpdateSeasonWeeks(db)) // correct the number of weeks

		// Participant Management
		commissioner.POST("/seasons/:season_id/participants", handlers.AddSeasonParticipants(db))              // add user(s) to a season
		commissioner.DELETE("/seasons/:season_id/participants/:user_id", handlers.RemoveSeasonParticipant(db)) // remove a user from a season

		// Week Management (manual steps only)
		commissioner.PUT("/weeks/:week_id/spreads", handlers.UpdateSpreads(db))                  // edit week spreads
		commissioner.POST("/weeks/:week_id/spreads/auto-import", handlers.AutoImportSpreads(db)) // auto-import spreads from Odds API
		commissioner.POST("/weeks/:week_id/activate", handlers.ActivateWeek(db))                 // activates a week

		// Pick Management
		commissioner.GET("/weeks/:week_id/picks", handlers.GetWeekPickSummary(db)) // returns user pick summaries for a week
	}

	// Admin-only routes
	admin := api.Group("/admin")
	admin.Use(middleware.RequireRole("admin"))
	{
		// admin.DELETE("/users/:id", handlers.DeleteUser(db))
		admin.GET("/users", handlers.GetAllAccounts(db))
		admin.PUT("/settings", handlers.UpdateGlobalSettings(db))
	}
}
