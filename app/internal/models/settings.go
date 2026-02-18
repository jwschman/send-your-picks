package models

import "time"

type Settings struct {
	ID string `db:"id" json:"id"`

	PickCutoffMinutes          int    `db:"pick_cutoff_minutes" json:"pick_cutoff_minutes"`
	AllowPickEdits             bool   `db:"allow_pick_edits" json:"allow_pick_edits"`
	PointsPerCorrectPick       int    `db:"points_per_correct_pick" json:"points_per_correct_pick"`
	CompetitionTimezone        string `db:"competition_timezone" json:"competition_timezone"`
	AllowCommissionerOverrides bool   `db:"allow_commissioner_overrides" json:"allow_commissioner_overrides"`
	AllowPicksAfterKickoff     bool   `db:"allow_picks_after_kickoff" json:"allow_picks_after_kickoff"`
	DebugMode                  bool   `db:"debug_mode" json:"debug_mode"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
