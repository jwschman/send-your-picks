package admin_bootstrap

import (
	"github.com/jmoiron/sqlx"
	"pawked.com/sendyourpicks/internal/id"
	"pawked.com/sendyourpicks/internal/logger"
	"pawked.com/sendyourpicks/internal/models"
)

func strPtr(s string) *string { return &s }

var seedTeams = []models.Team{
	{Name: "Cardinals", Abbreviation: "ARI", City: "Arizona", IsActive: true, LogoURL: strPtr("cardinals.png")},
	{Name: "Falcons", Abbreviation: "ATL", City: "Atlanta", IsActive: true, LogoURL: strPtr("falcons.png")},
	{Name: "Ravens", Abbreviation: "BAL", City: "Baltimore", IsActive: true, LogoURL: strPtr("ravens.png")},
	{Name: "Bills", Abbreviation: "BUF", City: "Buffalo", IsActive: true, LogoURL: strPtr("bills.png")},
	{Name: "Panthers", Abbreviation: "CAR", City: "Carolina", IsActive: true, LogoURL: strPtr("panthers.png")},
	{Name: "Bears", Abbreviation: "CHI", City: "Chicago", IsActive: true, LogoURL: strPtr("bears.png")},
	{Name: "Bengals", Abbreviation: "CIN", City: "Cincinnati", IsActive: true, LogoURL: strPtr("bengals.png")},
	{Name: "Browns", Abbreviation: "CLE", City: "Cleveland", IsActive: true, LogoURL: strPtr("browns.png")},
	{Name: "Cowboys", Abbreviation: "DAL", City: "Dallas", IsActive: true, LogoURL: strPtr("cowboys.png")},
	{Name: "Broncos", Abbreviation: "DEN", City: "Denver", IsActive: true, LogoURL: strPtr("broncos.png")},
	{Name: "Lions", Abbreviation: "DET", City: "Detroit", IsActive: true, LogoURL: strPtr("lions.png")},
	{Name: "Packers", Abbreviation: "GB", City: "Green Bay", IsActive: true, LogoURL: strPtr("packers.png")},
	{Name: "Texans", Abbreviation: "HOU", City: "Houston", IsActive: true, LogoURL: strPtr("texans.png")},
	{Name: "Colts", Abbreviation: "IND", City: "Indianapolis", IsActive: true, LogoURL: strPtr("colts.png")},
	{Name: "Jaguars", Abbreviation: "JAX", City: "Jacksonville", IsActive: true, LogoURL: strPtr("jaguars.png")},
	{Name: "Chiefs", Abbreviation: "KC", City: "Kansas City", IsActive: true, LogoURL: strPtr("chiefs.png")},
	{Name: "Raiders", Abbreviation: "LV", City: "Las Vegas", IsActive: true, LogoURL: strPtr("raiders.png")},
	{Name: "Chargers", Abbreviation: "LAC", City: "Los Angeles", IsActive: true, LogoURL: strPtr("chargers.png")},
	{Name: "Rams", Abbreviation: "LAR", City: "Los Angeles", IsActive: true, LogoURL: strPtr("rams.png")},
	{Name: "Dolphins", Abbreviation: "MIA", City: "Miami", IsActive: true, LogoURL: strPtr("dolphins.png")},
	{Name: "Vikings", Abbreviation: "MIN", City: "Minnesota", IsActive: true, LogoURL: strPtr("vikings.png")},
	{Name: "Patriots", Abbreviation: "NE", City: "New England", IsActive: true, LogoURL: strPtr("patriots.png")},
	{Name: "Saints", Abbreviation: "NO", City: "New Orleans", IsActive: true, LogoURL: strPtr("saints.png")},
	{Name: "Giants", Abbreviation: "NYG", City: "New York", IsActive: true, LogoURL: strPtr("giants.png")},
	{Name: "Jets", Abbreviation: "NYJ", City: "New York", IsActive: true, LogoURL: strPtr("jets.png")},
	{Name: "Eagles", Abbreviation: "PHI", City: "Philadelphia", IsActive: true, LogoURL: strPtr("eagles.png")},
	{Name: "Steelers", Abbreviation: "PIT", City: "Pittsburgh", IsActive: true, LogoURL: strPtr("steelers.png")},
	{Name: "49ers", Abbreviation: "SF", City: "San Francisco", IsActive: true, LogoURL: strPtr("49ers.png")},
	{Name: "Seahawks", Abbreviation: "SEA", City: "Seattle", IsActive: true, LogoURL: strPtr("seahawks.png")},
	{Name: "Buccaneers", Abbreviation: "TB", City: "Tampa Bay", IsActive: true, LogoURL: strPtr("buccaneers.png")},
	{Name: "Titans", Abbreviation: "TEN", City: "Tennessee", IsActive: true, LogoURL: strPtr("titans.png")},
	{Name: "Commanders", Abbreviation: "WSH", City: "Washington", IsActive: true, LogoURL: strPtr("commanders.png")},
}

func SeedTeams(db *sqlx.DB) error {
	var count int

	err := db.Get(&count, `SELECT COUNT(*) FROM public.teams`)
	if err != nil {
		logger.Error("Seed Teams--Error Getting teams")
		return err
	}

	// If teams already exist, do nothing
	if count > 0 {
		logger.Info("Teams info already exists")
		return nil
	}

	// begin a transaction
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, team := range seedTeams {
		teamID, err := id.New()
		if err != nil {
			logger.Error("Error creating Team ID: %s", err)
			return err
		}

		_, err = tx.Exec(
			`
			INSERT INTO public.teams (
				id,
				name,
				abbreviation,
				city,
				is_active,
				logo_url
			)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (abbreviation) DO NOTHING
			`,
			teamID,
			team.Name,
			team.Abbreviation,
			team.City,
			team.IsActive,
			team.LogoURL,
		)

		if err != nil {
			logger.Error("Error seeding team data: %s", err)
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		logger.Error("Error commiting teams to database: %s", err)
	}

	return nil
}
