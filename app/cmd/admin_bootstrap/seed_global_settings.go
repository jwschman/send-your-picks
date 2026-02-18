package admin_bootstrap

import (
	"github.com/jmoiron/sqlx"
	"pawked.com/sendyourpicks/internal/id"
	"pawked.com/sendyourpicks/internal/logger"
)

var globalSettingsSeed = struct {
	PickCutoffMinutes          int
	AllowPickEdits             bool
	PointsPerCorrectPick       int
	CompetitionTimezone        string
	AllowCommissionerOverrides bool
}{
	PickCutoffMinutes:          120,
	AllowPickEdits:             true,
	PointsPerCorrectPick:       1,
	CompetitionTimezone:        "Asia/Tokyo",
	AllowCommissionerOverrides: false,
}

func SeedGlobalSettings(db *sqlx.DB) error {
	var count int

	err := db.Get(
		&count,
		`select count(*) from public.settings`,
	)
	if err != nil {
		return err
	}

	if count > 0 {
		logger.Info("Global settings already present")
		return nil
	}

	settingsID, err := id.New()
	if err != nil {
		return err
	}

	logger.Info("Global settings table empty.  Seeding default values...")

	_, err = db.Exec(
		`
		insert into public.settings (
			id,
			pick_cutoff_minutes,
			allow_pick_edits,
			points_per_correct_pick,
			competition_timezone,
			allow_commissioner_overrides
		)
		values ($1, $2, $3, $4, $5, $6)
		`,
		settingsID,
		globalSettingsSeed.PickCutoffMinutes,
		globalSettingsSeed.AllowPickEdits,
		globalSettingsSeed.PointsPerCorrectPick,
		globalSettingsSeed.CompetitionTimezone,
		globalSettingsSeed.AllowCommissionerOverrides,
	)
	if err != nil {
		return err
	}

	logger.Info("Global settings database seeded")
	return nil
}
