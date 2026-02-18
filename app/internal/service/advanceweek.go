package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"pawked.com/sendyourpicks/internal/external"
	"pawked.com/sendyourpicks/internal/id"
	"pawked.com/sendyourpicks/internal/logger"
	"pawked.com/sendyourpicks/internal/models"
	"pawked.com/sendyourpicks/internal/settings"
)

// Week status constants for the state machine
const (
	StatusDraft                  = "draft"
	StatusGamesImported          = "games_imported"
	StatusSpreadsSet             = "spreads_set"
	StatusActive                 = "active"
	StatusPlayed                 = "played"
	StatusPicksResultsCalculated = "picks_results_calculated"
	StatusScored                 = "scored"
	StatusFinal                  = "final"
)

var (
	ErrManualActionRequired = errors.New("manual action required")
	ErrWeekAlreadyFinal     = errors.New("week is already final")
	ErrWaitingForGames      = errors.New("waiting for games to finish")
	ErrInvalidWeekStatus    = errors.New("invalid week status")
	ErrSeasonComplete       = errors.New("season complete: all weeks have already been played")
)

// Advances the week's state.  Called from inside AdvanceSeason
func AdvanceWeekState(ctx context.Context, db *sqlx.DB, week *models.Week, actingUserID string) error {

	for {
		switch week.Status {

		// week created, no games imported yet
		case StatusDraft:
			// Automated: Import games from external API
			logger.Debug("AdvanceWeekState: importing games for week", "week_number", week.Number)

			res, err := ImportGamesForWeek(ctx, db, week.ID, actingUserID)
			if err != nil {
				return err
			}

			logger.Debug("AdvanceWeekState: imported games, status now games_imported", "games_created", res.GamesCreated)
			// Continue to next iteration (will hit manual state)

		// games imported, waiting for spreads
		case StatusGamesImported:
			// Manual: Commissioner needs to set spreads
			logger.Debug("AdvanceWeekState: waiting for commissioner to set spreads", "week_number", week.Number)
			return ErrManualActionRequired

		// some or all spreads are set, but not activated yet
		case StatusSpreadsSet:
			// Manual: Commissioner needs to activate the week
			logger.Debug("AdvanceWeekState: waiting for commissioner to activate", "week_number", week.Number)
			return ErrManualActionRequired

		// week is active and released to users to make picks.  Scores can start to populate in this state
		case StatusActive:
			// Automated: Import scores from external API
			logger.Debug("AdvanceWeekState: importing scores for week", "week_number", week.Number)

			res, err := ImportScoresForWeek(ctx, db, week.ID)
			if err != nil {
				return err
			}

			logger.Debug("AdvanceWeekState: score import complete",
				"games_updated", res.GamesUpdated,
				"games_in_db", res.GamesInDB,
				"games_from_api", res.GamesFromAPI,
				"unmatched_count", len(res.UnmatchedGameIDs),
				"all_games_played", res.AllGamesPlayed)

			// If not all games are finished, stop and wait
			if !res.AllGamesPlayed {
				return ErrWaitingForGames
			}
			// Continue to next iteration

		// All games in the week have been played
		case StatusPlayed:
			// Automated: Calculate pick correctness
			logger.Debug("AdvanceWeekState: calculating pick results for week", "week_number", week.Number)

			res, err := CalculatePickResults(ctx, db, week.ID)
			if err != nil {
				return err
			}

			logger.Debug("AdvanceWeekState: processed games and picks", "games_processed", res.GamesProcessed, "picks_updated", res.PicksUpdated)
			// Continue to next iteration

		// PickResults have been calculated from the games
		case StatusPicksResultsCalculated:
			// Automated: Calculate week points and rankings
			logger.Debug("AdvanceWeekState: calculating week points for week", "week_number", week.Number)

			res, err := CalculateWeekPoints(ctx, db, week.ID)
			if err != nil {
				return err
			}

			logger.Debug("AdvanceWeekState: processed users", "users_processed", res.UsersProcessed)
			// Continue to next iteration

		// Points have been assigned
		case StatusScored:
			// Automated: Calculate season standings snapshot
			logger.Debug("AdvanceWeekState: calculating season snapshot for week", "week_number", week.Number)

			res, err := CalculateSeasonSnapshot(ctx, db, week.ID)
			if err != nil {
				return err
			}

			logger.Debug("AdvanceWeekState: processed users, week now final", "users_processed", res.UsersProcessed)
			// Continue to next iteration (will exit on StatusFinal)

		// Standings have been calculated and the week is final
		case StatusFinal:
			// Done - exit the loop
			logger.Debug("AdvanceWeekState: week is now final", "week_number", week.Number)
			return nil

		// something is wrong
		default:
			return ErrInvalidWeekStatus
		}

		// Reload week from DB to get updated status for next iteration
		if err := db.Get(week, `SELECT * FROM weeks WHERE id = $1`, week.ID); err != nil {
			return err
		}
	}
}

// create the next week for the season.  basically what i had before, just a function rather than method
func CreateNextWeekForSeason(db *sqlx.DB, seasonID, userID string) (*models.Week, error) {

	// start transaction
	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var lastWeek struct {
		Number int    `db:"number"`
		Status string `db:"status"`
	}

	// find the most recent week (with lock to prevent concurrent week creation)
	err = tx.Get(&lastWeek, `
		SELECT number, status
		FROM weeks
		WHERE season_id = $1
		ORDER BY number DESC
		LIMIT 1
		FOR UPDATE
	`, seasonID)

	nextWeekNumber := 1

	// if the last week exists, make sure it's "final"
	if err == nil {
		if lastWeek.Status != "final" {
			return nil, errors.New("previous week not final")
		}
		// the next week will be the previous week +1
		nextWeekNumber = lastWeek.Number + 1
		// if the error is anything but ErrNoRows, return.  Otherwise we'll just create week 1
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// just make sure that the new week isn't higher than the maximum number.  But this should never even hit because the season should be finalized
	var numberOfWeeks int
	err = tx.Get(&numberOfWeeks, `SELECT number_of_weeks FROM seasons WHERE id = $1`, seasonID)
	if err != nil {
		return nil, err
	}

	if nextWeekNumber > numberOfWeeks {
		return nil, ErrSeasonComplete
	}

	// generate the new week ID
	weekID, err := id.New()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`
		INSERT INTO weeks (id, season_id, number, status, created_by)
		VALUES ($1, $2, $3, 'draft', $4)
	`, weekID, seasonID, nextWeekNumber, userID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	logger.Debug("New week created", "week_number", nextWeekNumber, "week_id", weekID)

	return &models.Week{
		ID:       weekID,
		SeasonID: seasonID,
		Number:   nextWeekNumber,
		Status:   "draft",
	}, nil
}

// this is going to be the fun one
type ImportGamesResult struct {
	GamesCreated int
}

var (
	ErrWeekNotFound      = errors.New("week not found")
	ErrInvalidWeekState  = errors.New("week not in draft state")
	ErrTeamMappingFailed = errors.New("team mapping failed")
)

// imports games from the external nfl client
func ImportGamesForWeek(ctx context.Context, db *sqlx.DB, weekID string, actingUserID string) (*ImportGamesResult, error) {

	// get the week with the year included
	week, err := GetWeekWithYear(db, weekID)
	if err != nil {
		return nil, err
	}

	logger.Debug("Beginning import for week", "week_number", week.Number, "week_id", week.ID)

	// make sure the week status is "draft"
	if week.Status != "draft" {
		return nil, ErrInvalidWeekState
	}

	// load the external client
	client, err := external.NewBallDontLieClient()
	if err != nil {
		return nil, err
	}

	externalGames, err := client.FetchGames(
		ctx,
		week.Year,
		week.Number,
		week.IsPostseason,
	)
	if err != nil {
		return nil, err
	}

	if len(externalGames) == 0 {
		return nil, fmt.Errorf("external API returned 0 games for week %d (year %d, postseason=%v)", week.Number, week.Year, week.IsPostseason)
	}

	// games from the external api validated and using the internal IDs and Abbrs
	type validatedGame struct {
		ExternalID   int64
		KickoffTime  time.Time
		HomeTeamID   string
		AwayTeamID   string
		HomeTeamAbbr string
		AwayTeamAbbr string
	}

	validated := make([]validatedGame, 0)

	// go through the external games and turn them into validatedGames
	for _, g := range externalGames {
		homeTeamID, err := GetTeamIDByAbbreviation(db, g.HomeTeam.Abbreviation)
		if err != nil {
			return nil, err
		}

		awayTeamID, err := GetTeamIDByAbbreviation(db, g.AwayTeam.Abbreviation)
		if err != nil {
			return nil, err
		}

		// add it to the list
		validated = append(validated, validatedGame{
			ExternalID:   g.ID,
			KickoffTime:  g.Date,
			HomeTeamID:   homeTeamID,
			AwayTeamID:   awayTeamID,
			HomeTeamAbbr: g.HomeTeam.Abbreviation,
			AwayTeamAbbr: g.AwayTeam.Abbreviation,
		})
	}

	// start the transaction
	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// always make sure there are no games whenever we import a week
	if _, err := tx.Exec(
		`DELETE FROM public.games WHERE week_id = $1`,
		weekID,
	); err != nil {
		return nil, err
	}

	created := 0

	// just loop through all the games
	for _, g := range validated {
		// we need to generate a new id for each game
		gameID, err := id.New()
		if err != nil {
			return nil, err
		}

		if _, err := tx.Exec(`
			INSERT INTO public.games (
				id,
				week_id,
				season_id,
				external_game_id,
				kickoff_time,
				home_team_id,
				away_team_id,
				home_team_abbr,
				away_team_abbr,
				created_by,
				home_spread,
				neutral_site,
				status
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NULL,FALSE,$11)
		`,
			gameID,
			weekID,
			week.SeasonID,
			g.ExternalID,
			g.KickoffTime,
			g.HomeTeamID,
			g.AwayTeamID,
			g.HomeTeamAbbr,
			g.AwayTeamAbbr,
			actingUserID,
			"scheduled",
		); err != nil {
			return nil, err
		}

		created++
	}

	// update week status to games_imported
	if err := UpdateWeekStatus(tx, weekID, "games_imported"); err != nil {
		return nil, err
	}

	// commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	logger.Debug("importGames: completed successfully", "week_id", week.ID, "games_created", created, "status", "games_imported")

	return &ImportGamesResult{
		GamesCreated: created,
	}, nil
}

type ImportScoresResult struct {
	GamesUpdated     int
	AllGamesPlayed   bool
	GamesInDB        int     // total games in our database for this week
	GamesFromAPI     int     // total games returned by external API
	UnmatchedGameIDs []int64 // external game IDs that didn't match any game in our DB
}

var (
	ErrWeekNotActive = errors.New("week not in active state")
)

// ImportScoresForWeek fetches scores from the external API and updates games.
// If all games become final, it transitions the week to "played" status.
func ImportScoresForWeek(ctx context.Context, db *sqlx.DB, weekID string) (*ImportScoresResult, error) {

	week, err := GetWeekWithYear(db, weekID)
	if err != nil {
		return nil, err
	}

	// make sure the week's status is "active"
	if week.Status != "active" {
		return nil, ErrWeekNotActive
	}

	logger.Debug("ImportScoresForWeek: fetching scores", "week_number", week.Number, "postseason", week.IsPostseason, "year", week.Year)

	// start up the external nfl client and get the games
	client, err := external.NewBallDontLieClient()
	if err != nil {
		return nil, err
	}

	externalGames, err := client.FetchGames(ctx, week.Year, week.Number, week.IsPostseason)
	if err != nil {
		return nil, err
	}

	// set up the transaction
	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// count total games in our database for this week
	var gamesInDB int
	if err := tx.Get(&gamesInDB, `SELECT COUNT(*) FROM public.games WHERE week_id = $1`, weekID); err != nil {
		return nil, err
	}

	// track updated games and unmatched external games
	updated := 0
	var unmatchedGameIDs []int64

	for _, externalGame := range externalGames {
		// only import final games for now (in the future i'd like to import game scores as they go, but that's kind of hard to do in the offseason)
		if externalGame.Status != "Final" && externalGame.Status != "Final/OT" {
			continue
		}

		result, err := tx.Exec(`
			UPDATE public.games
			SET
				home_score = $1,
				away_score = $2,
				status = 'final',
				updated_at = NOW()
			WHERE external_game_id = $3
			AND week_id = $4
		`, externalGame.HomeTeamScore, externalGame.AwayTeamScore, externalGame.ID, weekID)
		if err != nil {
			return nil, err
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			// this external game didn't match any game in our database
			unmatchedGameIDs = append(unmatchedGameIDs, externalGame.ID)
		} else {
			updated += int(rows)
		}
	}

	// check if all games are now final
	var gamesNotFinal int
	err = tx.Get(&gamesNotFinal, `
		SELECT COUNT(*)
		FROM public.games
		WHERE week_id = $1 AND status <> 'final'
	`, weekID)
	if err != nil {
		return nil, err
	}

	allGamesPlayed := gamesNotFinal == 0

	// if all games are final, update week status to played
	if allGamesPlayed {
		if err := UpdateWeekStatus(tx, weekID, "played"); err != nil {
			return nil, err
		}
		logger.Debug("ImportScoresForWeek: all games final, week status set to played")
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	if len(unmatchedGameIDs) > 0 {
		logger.Warn("ImportScoresForWeek: some external games did not match any game in database",
			"unmatched_count", len(unmatchedGameIDs),
			"unmatched_ids", unmatchedGameIDs)
	}

	logger.Debug("ImportScoresForWeek: completed",
		"games_updated", updated,
		"games_in_db", gamesInDB,
		"games_from_api", len(externalGames),
		"all_games_played", allGamesPlayed)

	return &ImportScoresResult{
		GamesUpdated:     updated,
		AllGamesPlayed:   allGamesPlayed,
		GamesInDB:        gamesInDB,
		GamesFromAPI:     len(externalGames),
		UnmatchedGameIDs: unmatchedGameIDs,
	}, nil
}

///////////////////////////////////
// stuff for calculate pick results
///////////////////////////////////

// CalculatePickResults determines which picks were correct based on game results and spreads.
// Transitions week from "played" to "picks_results_calculated".

// this struct name sucks
type CalculatePickResultsResult struct {
	GamesProcessed int
	PicksUpdated   int
}

var (
	ErrWeekNotPlayed = errors.New("week not in played state")
)

// WinningTeamByGame returns the winning team ID based on score + spread, and nil for ties.
func WinningTeamByGame(game models.Game) *string {

	// hometeam score + the spread (which can be minus)
	adjustedHomeScore := float64(*game.HomeScore) + *game.HomeSpread
	awayScore := float64(*game.AwayScore)

	//home team wins
	if adjustedHomeScore > awayScore {
		return &game.HomeTeamID
	}
	// away team wins
	if adjustedHomeScore < awayScore {
		return &game.AwayTeamID
	}
	// tie
	return nil
}

// calculates if a pick was correct or not
func CalculatePickResults(ctx context.Context, db *sqlx.DB, weekID string) (*CalculatePickResultsResult, error) {
	// all we need here is just the week's status, not the whole week
	weekStatus, err := GetWeekStatus(db, weekID)
	if err != nil {
		return nil, err
	}
	if weekStatus != "played" {
		return nil, ErrWeekNotPlayed
	}

	logger.Debug("CalculatePickResults: starting for week", "week_id", weekID)

	games, err := GetWeekGames(db, weekID)
	if err != nil {
		return nil, err
	}

	picks, err := GetWeekPicks(db, weekID)
	if err != nil {
		return nil, err
	}

	// group picks by game ID
	picksByGameID := make(map[string][]models.Pick)
	for _, pick := range picks {
		picksByGameID[pick.GameID] = append(picksByGameID[pick.GameID], pick)
	}

	// set up transaction
	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// use a single now
	now := time.Now().UTC()

	// initialize the amount of games and picks processed
	gamesProcessed := 0
	picksUpdated := 0

	// range over the games and determine the results
	for _, game := range games {
		// first find out the ID of the team that won
		winningTeamID := WinningTeamByGame(game)

		for _, pick := range picksByGameID[game.ID] {
			var isCorrect *bool

			if pick.SelectedTeamID != nil && winningTeamID != nil {
				// pointers are fun
				value := *pick.SelectedTeamID == *winningTeamID
				isCorrect = &value
			}

			// picks with no selected_team_id as well as draw games will get NULL for is_correct, but will still get a calculated_at timestamp
			query := `
				UPDATE public.picks
				SET is_correct = $1,
					calculated_at = $2
				WHERE id = $3
			`
			_, err = tx.Exec(query, isCorrect, now, pick.ID)
			if err != nil {
				return nil, err
			}

			picksUpdated++
		}

		gamesProcessed++
	}

	// update week status to picks_results_calculated
	// should this be a separate function as well?
	if err := UpdateWeekStatus(tx, weekID, "picks_results_calculated"); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	logger.Debug("CalculatePickResults: completed", "games_processed", gamesProcessed, "picks_updated", picksUpdated)

	return &CalculatePickResultsResult{
		GamesProcessed: gamesProcessed,
		PicksUpdated:   picksUpdated,
	}, nil
}

// CalculateWeekPoints calculates points and rankings for all users in a week
// Also transitions week from "picks_results_calculated" to "scored"
type CalculateWeekPointsResult struct {
	UsersProcessed int
}

var (
	ErrWeekNotPicksCalculated = errors.New("week not in picks_results_calculated state")
)

// calculates the weeks points and stores the results in week_results
func CalculateWeekPoints(ctx context.Context, db *sqlx.DB, weekID string) (*CalculateWeekPointsResult, error) {
	// Get week info including season_id (needed for participant filtering)
	week, err := GetWeekWithYear(db, weekID)
	if err != nil {
		return nil, err
	}
	if week.Status != "picks_results_calculated" {
		return nil, ErrWeekNotPicksCalculated
	}

	logger.Debug("CalculateWeekPoints: starting for week", "week_id", weekID)

	// get settings
	s, err := settings.Get(db)
	if err != nil {
		return nil, err
	}

	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// calculation time
	now := time.Now().UTC()

	// Calculate points for each season participant.
	// Only participants are included - non-participants don't get week_results entries.
	// Users with no picks for the week get 0 points (LEFT JOIN on picks).
	type UserPoints struct {
		UserID string `db:"id"`
		Points int    `db:"points"`
	}

	var userPoints []UserPoints
	err = tx.Select(&userPoints, `
		SELECT
			p.id,
			COALESCE(COUNT(*) FILTER (WHERE pk.is_correct = true) * $1, 0) as points
		FROM public.season_participants sp
		JOIN public.profiles p ON p.id = sp.user_id
		LEFT JOIN public.picks pk ON pk.user_id = p.id AND pk.week_id = $2
		WHERE sp.season_id = $3
		GROUP BY p.id
	`, s.PointsPerCorrectPick, weekID, week.SeasonID)
	if err != nil {
		return nil, err
	}

	// insert or update week_results
	usersProcessed := 0
	for _, up := range userPoints {

		// we always generate a new ID.  It will be tossed if the row already exists, but whatever.
		resultID, err := id.New()
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec(`
			INSERT INTO public.week_results
			(id, user_id, week_id, points, computed_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (user_id, week_id)
			DO UPDATE SET
				points = EXCLUDED.points,
				computed_at = EXCLUDED.computed_at,
				updated_at = EXCLUDED.updated_at
		`, resultID, up.UserID, weekID, up.Points, now, now, now)
		if err != nil {
			return nil, err
		}
		usersProcessed++
	}

	// calculate and update ranks
	_, err = tx.Exec(`
		UPDATE public.week_results
		SET rank = subquery.rank
		FROM (
			SELECT
				id,
				RANK() OVER (ORDER BY points DESC) as rank
			FROM public.week_results
			WHERE week_id = $1
		) subquery
		WHERE week_results.id = subquery.id
	`, weekID)
	if err != nil {
		return nil, err
	}

	// update week status to scored
	if err := UpdateWeekStatus(tx, weekID, "scored"); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	logger.Debug("CalculateWeekPoints: completed", "users_processed", usersProcessed)

	return &CalculateWeekPointsResult{
		UsersProcessed: usersProcessed,
	}, nil
}

// CalculateSeasonSnapshot calculates cumulative standings for the season after a week.
// Transitions week from "scored" to "final".
type CalculateSeasonSnapshotResult struct {
	UsersProcessed int
}

var (
	ErrWeekNotScored = errors.New("week not in scored state")
)

func CalculateSeasonSnapshot(ctx context.Context, db *sqlx.DB, weekID string) (*CalculateSeasonSnapshotResult, error) {
	// we need specifically the season id and status
	week, err := GetWeekWithYear(db, weekID)

	if err != nil {
		return nil, err
	}
	if week.Status != "scored" {
		return nil, ErrWeekNotScored
	}

	logger.Debug("CalculateSeasonSnapshot: starting", "week_id", weekID, "season_id", week.SeasonID)

	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// load this week's results
	type WeekResult struct {
		UserID string `db:"user_id"`
		Points int    `db:"points"`
	}
	var weekResults []WeekResult
	err = tx.Select(&weekResults, `SELECT user_id, points FROM public.week_results WHERE week_id = $1`, weekID)
	if err != nil {
		return nil, err
	}

	// load previous week's standings (will be empty for week 1)
	var previousStandings []WeekResult
	err = tx.Select(&previousStandings, `
		SELECT user_id, points
		FROM public.season_standings
		WHERE season_id = $1
		AND week_id = (
			SELECT id
			FROM public.weeks
			WHERE season_id = $1
			AND number < (
				SELECT number
				FROM public.weeks
				WHERE id = $2
			)
			ORDER BY number DESC
			LIMIT 1
		)
	`, week.SeasonID, weekID)
	if err != nil {
		return nil, err
	}

	// create map of previous points
	previousPointsByUserID := make(map[string]int)
	for _, standing := range previousStandings {
		previousPointsByUserID[standing.UserID] = standing.Points
	}

	// initialize time and number of users processed
	now := time.Now().UTC()
	usersProcessed := 0

	// insert/update standings for each user
	for _, weekResult := range weekResults {
		standingID, err := id.New()
		if err != nil {
			return nil, err
		}

		previousPoints := previousPointsByUserID[weekResult.UserID]
		cumulativePoints := previousPoints + weekResult.Points

		_, err = tx.Exec(`
			INSERT INTO public.season_standings
			(id, user_id, season_id, week_id, points, computed_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (user_id, season_id, week_id)
			DO UPDATE SET
				points = EXCLUDED.points,
				computed_at = EXCLUDED.computed_at,
				updated_at = EXCLUDED.updated_at
		`, standingID, weekResult.UserID, week.SeasonID, weekID, cumulativePoints, now, now, now)
		if err != nil {
			return nil, err
		}
		usersProcessed++
	}

	// calculate and update ranks
	_, err = tx.Exec(`
		UPDATE public.season_standings
		SET rank = subquery.rank
		FROM (
			SELECT
				id,
				RANK() OVER (ORDER BY points DESC) as rank
			FROM public.season_standings
			WHERE season_id = $1 AND week_id = $2
		) subquery
		WHERE season_standings.id = subquery.id
	`, week.SeasonID, weekID)
	if err != nil {
		return nil, err
	}

	// update week status to final and set closed_at
	if _, err := tx.Exec(`UPDATE public.weeks SET status = 'final', closed_at = NOW(), updated_at = NOW() WHERE id = $1`, weekID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	logger.Debug("CalculateSeasonSnapshot: completed", "users_processed", usersProcessed, "status", "final")

	return &CalculateSeasonSnapshotResult{
		UsersProcessed: usersProcessed,
	}, nil
}
