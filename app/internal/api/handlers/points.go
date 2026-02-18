// SPECIFICALLY CALL EVERYTHING HERE POINTS
// SCORES ARE ON NFL GAMES
// POINTS ARE WHAT PLAYERS GET

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"pawked.com/sendyourpicks/internal/api/middleware"
	"pawked.com/sendyourpicks/internal/models"
	"pawked.com/sendyourpicks/internal/service"
)

// StandingWithUser is used for returning standings with username
type StandingWithUser struct {
	UserID   string `json:"user_id" db:"user_id"`
	Username string `json:"username" db:"username"`
	Points   int    `json:"points" db:"points"`
	Rank     int    `json:"rank" db:"rank"`
}

// GetWeekResults returns a week's results for all users
func GetWeekResults(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		weekID := c.Param("week_id")

		// make sure the week exists
		weekExists, err := service.WeekExists(db, weekID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database Error",
			})
			return
		} else if !weekExists {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Week not found",
				"week_id": weekID,
			})
			return
		}

		// initially I just returned the user ID, and that didn't have usernames
		type WeekResultWithUser struct {
			ID       string `json:"id" db:"id"`
			UserID   string `json:"user_id" db:"user_id"`
			Points   int    `json:"points" db:"points"`
			Rank     int    `json:"rank" db:"rank"`
			Username string `json:"username" db:"username"`
		}

		// Query week results, filtered to only include season participants.
		// Joins through weeks to get season_id for the participant check.
		var weekResults []WeekResultWithUser
		query := `
			SELECT
				wr.id,
				wr.user_id,
				wr.points,
				wr.rank,
				p.username
			FROM public.week_results wr
			JOIN public.profiles p ON p.id = wr.user_id
			JOIN public.weeks w ON w.id = wr.week_id
			JOIN public.season_participants sp ON sp.season_id = w.season_id AND sp.user_id = wr.user_id
			WHERE wr.week_id = $1
			ORDER BY wr.rank ASC
		`

		err = db.Select(&weekResults, query, weekID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database Error",
			})
			return
		}

		if weekResults == nil {
			weekResults = []WeekResultWithUser{}
		}

		c.JSON(http.StatusOK, gin.H{
			"week_id":      weekID,
			"week_results": weekResults,
		})

	}
}

// WeekPointsByUserID returns how many points each user got for the given week
func WeekPointsByUserID(results []models.WeekResult) map[string]int {
	weekPointsByUserID := make(map[string]int)

	for _, r := range results {
		weekPointsByUserID[r.UserID] = r.Points
	}

	return weekPointsByUserID
}

// CumulativePointsByUserID returns cumulative points a user has
func CumulativePointsByUserID(standing []models.SeasonStanding) map[string]int {
	cumulativePointsByUserID := make(map[string]int)

	for _, s := range standing {
		cumulativePointsByUserID[s.UserID] = s.Points
	}
	return cumulativePointsByUserID
}

// Returns the standings for all users after a given week
func GetSeasonStandings(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		weekID := c.Param("week_id")

		// make sure the week exists
		weekExists, err := service.WeekExists(db, weekID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database Error",
			})
			return
		} else if !weekExists {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Week not found",
				"week_id": weekID,
			})
			return
		}

		// Query season standings after a given week, filtered to only include participants.
		// Joins through weeks to get season_id for the participant check.
		var standings []StandingWithUser
		query := `
			SELECT
				ss.user_id,
				ss.points,
				ss.rank,
				p.username
			FROM public.season_standings ss
			JOIN public.profiles p ON p.id = ss.user_id
			JOIN public.weeks w ON w.id = ss.week_id
			JOIN public.season_participants sp ON sp.season_id = w.season_id AND sp.user_id = ss.user_id
			WHERE ss.week_id = $1
			ORDER BY ss.rank ASC
		`
		err = db.Select(&standings, query, weekID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting standings"})
			return
		}

		if standings == nil {
			standings = []StandingWithUser{}
		}

		c.JSON(http.StatusOK, gin.H{
			"standings": standings,
			"week_id":   weekID,
		})
	}
}

// returns the latest standings for all users for the requested season
func GetCurrentSeasonStandings(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		seasonID := c.Param("season_id")

		// make sure the season exists
		seasonExists, err := service.SeasonExists(db, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database Error",
			})
			return
		} else if !seasonExists {
			c.JSON(http.StatusNotFound, gin.H{
				"error":     "Season not found",
				"season_id": seasonID,
			})
			return
		}

		// Query latest standings for the season, filtered to only include participants.
		var latestStandings []StandingWithUser
		query := `
			SELECT
				ss.user_id,
				ss.points,
				ss.rank,
				p.username
			FROM public.season_standings ss
			JOIN public.profiles p ON p.id = ss.user_id
			JOIN public.season_participants sp ON sp.season_id = $1 AND sp.user_id = ss.user_id
			WHERE ss.week_id = (
				SELECT id
				FROM public.weeks
				WHERE season_id = $1
				AND status = 'final'
				ORDER BY number DESC
				LIMIT 1
			)
			ORDER BY ss.rank ASC
		`
		err = db.Select(&latestStandings, query, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting standings"})
			return
		}

		if latestStandings == nil {
			latestStandings = []StandingWithUser{}
		}

		c.JSON(http.StatusOK, gin.H{
			"standings": latestStandings,
			"season_id": seasonID,
		})
	}
}

// GetSeasonPoints returns how many points a user got per week in the season
func GetMySeasonPoints(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		seasonID := c.Param("season_id")

		// make sure the week exists
		seasonExists, err := service.SeasonExists(db, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database Error",
			})
			return
		} else if !seasonExists {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Week not found",
				"week_id": seasonID,
			})
			return
		}

		// initially I just returned the user ID, and that didn't have usernames
		type WeeksWithPoints struct {
			WeekNumber  int `json:"week_number" db:"week_number"`
			WeekPoints  int `json:"week_points" db:"week_points"`
			WeekRank    int `json:"week_rank" db:"week_rank"`
			TotalPoints int `json:"total_points" db:"total_points"`
			LeagueRank  int `json:"league_rank" db:"league_rank"`
		}

		// query by ChatGPT
		var weeksWithPoints []WeeksWithPoints
		query := `
					SELECT
						w.number                         AS week_number,
						COALESCE(wr.points, 0)           AS week_points,
						wr.rank                          AS week_rank,
						ss.points                        AS total_points,
						ss.rank                          AS league_rank
					FROM public.weeks w
					LEFT JOIN public.week_results wr
						ON wr.week_id = w.id
					AND wr.user_id = $2
					LEFT JOIN public.season_standings ss
						ON ss.week_id = w.id
					AND ss.user_id = $2
					WHERE w.season_id = $1
					AND w.status = 'final' 
					ORDER BY w.number ASC;
				`

		err = db.Select(&weeksWithPoints, query, seasonID, userID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database Error",
			})
			return
		}

		if weeksWithPoints == nil {
			weeksWithPoints = []WeeksWithPoints{}
		}

		c.JSON(http.StatusOK, gin.H{
			"season_id":         seasonID,
			"weeks_with_points": weeksWithPoints,
		})

	}
}

// GetWeekWinners returns who won each week of a season (with tie info)
func GetWeekWinners(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		seasonID := c.Param("season_id")

		seasonExists, err := service.SeasonExists(db, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database Error"})
			return
		} else if !seasonExists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Season not found", "season_id": seasonID})
			return
		}

		type FirstPlaceFinish struct {
			WeekID       string  `db:"week_id"`
			WeekNumber   int     `db:"week_number"`
			Points       int     `db:"points"`
			UserID       string  `db:"user_id"`
			Username     string  `db:"username"`
			AvatarURL    *string `db:"avatar_url"`
			WinnersCount int     `db:"winners_count"`
		}

		// Query week winners, filtered to only include season participants.
		var finishes []FirstPlaceFinish
		query := `
			SELECT
				w.id as week_id,
				w.number as week_number,
				wr.points,
				wr.user_id,
				p.username,
				p.avatar_url,
				COUNT(*) OVER (PARTITION BY w.id) as winners_count
			FROM public.weeks w
			JOIN public.week_results wr ON wr.week_id = w.id AND wr.rank = 1
			JOIN public.profiles p ON p.id = wr.user_id
			JOIN public.season_participants sp ON sp.season_id = $1 AND sp.user_id = wr.user_id
			WHERE w.season_id = $1
			AND w.status = 'final'
			ORDER BY w.number ASC
		`

		err = db.Select(&finishes, query, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting week winners"})
			return
		}

		type Winner struct {
			UserID    string `json:"user_id"`
			Username  string `json:"username"`
			AvatarURL string `json:"avatar_url"`
		}
		type WeekWinners struct {
			WeekID     string   `json:"week_id"`
			WeekNumber int      `json:"week_number"`
			Points     int      `json:"points"`
			Winners    []Winner `json:"winners"`
			IsTie      bool     `json:"is_tie"`
		}

		weekMap := make(map[string]*WeekWinners)
		var weekOrder []string

		for _, f := range finishes {
			if _, exists := weekMap[f.WeekID]; !exists {
				weekMap[f.WeekID] = &WeekWinners{
					WeekID:     f.WeekID,
					WeekNumber: f.WeekNumber,
					Points:     f.Points,
					Winners:    []Winner{},
					IsTie:      f.WinnersCount > 1,
				}
				weekOrder = append(weekOrder, f.WeekID)
			}
			weekMap[f.WeekID].Winners = append(weekMap[f.WeekID].Winners, Winner{
				UserID:    f.UserID,
				Username:  f.Username,
				AvatarURL: buildAvatarURL(f.AvatarURL),
			})
		}

		weeks := make([]WeekWinners, 0, len(weekOrder))
		for _, weekID := range weekOrder {
			weeks = append(weeks, *weekMap[weekID])
		}

		c.JSON(http.StatusOK, gin.H{
			"season_id": seasonID,
			"weeks":     weeks,
		})
	}
}

// GetUserWinCounts returns how many weeks each user has won (with tie breakdown)
func GetUserWinCounts(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		seasonID := c.Param("season_id")

		seasonExists, err := service.SeasonExists(db, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database Error"})
			return
		} else if !seasonExists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Season not found", "season_id": seasonID})
			return
		}

		type FirstPlaceFinish struct {
			UserID       string  `db:"user_id"`
			Username     string  `db:"username"`
			AvatarURL    *string `db:"avatar_url"`
			WinnersCount int     `db:"winners_count"`
		}

		// Query first-place finishes, filtered to only include season participants.
		var finishes []FirstPlaceFinish
		query := `
			SELECT
				wr.user_id,
				p.username,
				p.avatar_url,
				COUNT(*) OVER (PARTITION BY w.id) as winners_count
			FROM public.weeks w
			JOIN public.week_results wr ON wr.week_id = w.id AND wr.rank = 1
			JOIN public.profiles p ON p.id = wr.user_id
			JOIN public.season_participants sp ON sp.season_id = $1 AND sp.user_id = wr.user_id
			WHERE w.season_id = $1
			AND w.status = 'final'
		`

		err = db.Select(&finishes, query, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting win counts"})
			return
		}

		type UserWins struct {
			UserID    string `json:"user_id"`
			Username  string `json:"username"`
			AvatarURL string `json:"avatar_url"`
			Wins      int    `json:"wins"` // outright wins
			Ties      int    `json:"ties"` // tied for first
		}

		userMap := make(map[string]*UserWins)
		for _, f := range finishes {
			if _, exists := userMap[f.UserID]; !exists {
				userMap[f.UserID] = &UserWins{
					UserID:    f.UserID,
					Username:  f.Username,
					AvatarURL: buildAvatarURL(f.AvatarURL),
				}
			}
			if f.WinnersCount == 1 {
				userMap[f.UserID].Wins++
			} else {
				userMap[f.UserID].Ties++
			}
		}

		users := make([]UserWins, 0, len(userMap))
		for _, u := range userMap {
			users = append(users, *u)
		}

		// Sort by wins descending, then by ties descending
		for i := 0; i < len(users)-1; i++ {
			for j := i + 1; j < len(users); j++ {
				if users[j].Wins > users[i].Wins ||
					(users[j].Wins == users[i].Wins && users[j].Ties > users[i].Ties) {
					users[i], users[j] = users[j], users[i]
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"season_id": seasonID,
			"users":     users,
		})
	}
}

// returns the users current season standings
func GetMyCurrentSeasonStandings(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		seasonID := c.Param("season_id")

		// make sure the season exists
		seasonExists, err := service.SeasonExists(db, seasonID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database Error"})
			return
		} else if !seasonExists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Season not found", "season_id": seasonID})
			return
		}

		// Check if user is a participant in this season.
		// Non-participants get 404 since they don't have standings.
		var isParticipant bool
		err = db.Get(&isParticipant, `
			SELECT EXISTS (
				SELECT 1 FROM public.season_participants
				WHERE season_id = $1 AND user_id = $2
			)
		`, seasonID, userID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database Error"})
			return
		}
		if !isParticipant {
			c.JSON(http.StatusNotFound, gin.H{"error": "You are not a participant in this season"})
			return
		}

		// get the latest standings with username
		var myCurrentStandings struct {
			UserID     string `json:"user_id" db:"user_id"`
			Username   string `json:"username" db:"username"`
			Points     int    `json:"points" db:"points"`
			Rank       *int   `json:"rank" db:"rank"` // Pointer to handle NULL
			TotalUsers int    `json:"total_users" db:"total_users"`
			SeasonID   string `json:"season_id" db:"season_id"`
		}

		// Query to get user's current standings.
		// total_users counts only season participants, not all profiles.
		query := `
			WITH total_participants AS (
				SELECT COUNT(*) as total
				FROM public.season_participants
				WHERE season_id = $1
			),
			latest_standings AS (
				SELECT
					user_id,
					points,
					rank
				FROM public.season_standings
				WHERE season_id = $1
					AND user_id = $2
				ORDER BY created_at DESC
				LIMIT 1
			)
			SELECT
				p.id as user_id,
				p.username,
				COALESCE(ls.points, 0) as points,
				ls.rank,
				tp.total as total_users,
				$1 as season_id
			FROM public.profiles p
			CROSS JOIN total_participants tp
			LEFT JOIN latest_standings ls ON ls.user_id = p.id
			WHERE p.id = $2
		`

		err = db.Get(&myCurrentStandings, query, seasonID, userID)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting standings"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"my_standings": myCurrentStandings})
	}
}
