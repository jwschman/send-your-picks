package settings

import (
	"github.com/jmoiron/sqlx"
)

type Settings struct {
	PickCutoffMinutes          int    `db:"pick_cutoff_minutes"`
	AllowPickEdits             bool   `db:"allow_pick_edits"`
	PointsPerCorrectPick       int    `db:"points_per_correct_pick"`
	CompetitionTimezone        string `db:"competition_timezone"`
	AllowCommissionerOverrides bool   `db:"allow_commissioner_overrides"`
	AllowPicksAfterKickoff     bool   `db:"allow_picks_after_kickoff"`
}

// just gets all the global settings
func Get(db *sqlx.DB) (*Settings, error) {
	var s Settings
	err := db.Get(&s, `SELECT pick_cutoff_minutes, allow_pick_edits, points_per_correct_pick, competition_timezone, allow_commissioner_overrides, allow_picks_after_kickoff FROM settings LIMIT 1`)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
