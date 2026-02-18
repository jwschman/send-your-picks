// Core domain types for Send Your Picks

export type Season = {
	id: string;
	year: number;
	is_active: boolean;
	is_postseason: boolean;
	number_of_weeks: number;
	weeks?: Week[];
};

export type Week = {
	id: string;
	season_id: string;
	number: number;
	status: WeekStatus;
	activated_at: string | null;
	closed_at: string | null;
	created_at: string;
	updated_at: string;
	created_by: string;
	year?: number;
	is_postseason?: boolean;
	games: Game[];
};

export type WeekStatus =
	| 'draft'
	| 'games_imported'
	| 'spreads_set'
	| 'active'
	| 'played'
	| 'picks_results_calculated'
	| 'scored'
	| 'final';

export type Game = {
	id: string;
	external_game_id: string;
	season_id: string;
	week_id: string;
	home_team_id: string;
	away_team_id: string;
	home_score: number | null;
	away_score: number | null;
	home_spread: number | null;
	kickoff_time: string;
	status: string;
	neutral_site: boolean;
	created_at: string;
	updated_at: string;
	created_by: string;
	home_team_name: string;
	home_team_city: string;
	home_team_abbr: string;
	home_team_logo_url: string;
	away_team_name: string;
	away_team_city: string;
	away_team_abbr: string;
	away_team_logo_url: string;
};

export type Pick = {
	id: string;
	user_id: string;
	week_id: string;
	game_id: string;
	selected_team_id: string | null;
	is_correct: boolean | null;
	calculated_at: string | null;
	user_locked_at: string | null;
	created_at: string;
	updated_at: string;
};

export type Standing = {
	user_id: string;
	username: string;
	points: number;
	rank: number;
	total_users?: number;
	season_id?: string;
};

export type PublicProfile = {
	id: string;
	username: string | null;
	tagline: string | null;
	role: string;
	avatar_url: string | null;
};

export type User = {
	id: string;
	username: string | null;
	tagline: string | null;
	role: string;
	avatar_url: string | null;
	email: string;
	last_sign_in_at: string | null;
	created_at: string;
	updated_at: string;
};

export type WeekResult = {
	id: string;
	user_id: string;
	points: number;
	rank: number;
	username: string;
};

export type Settings = {
	id: string;
	pick_cutoff_minutes: number;
	allow_pick_edits: boolean;
	points_per_correct_pick: number;
	competition_timezone: string;
	allow_commissioner_overrides: boolean;
	allow_picks_after_kickoff: boolean;
	debug_mode: boolean;
	created_at?: string;
	updated_at?: string;
};

export type WeekWinner = {
	user_id: string;
	username: string;
	avatar_url: string;
};

export type WeekWinnersData = {
	week_id: string;
	week_number: number;
	points: number;
	winners: WeekWinner[];
	is_tie: boolean;
};

export type UserWinCount = {
	user_id: string;
	username: string;
	avatar_url: string;
	wins: number;
	ties: number;
};

export type Badge = {
	type: 'previous_week_winner' | 'previous_season_winner' | 'previous_season_loser';
	label: string;
};

export type StandingsHistoryEntry = {
	user_id: string;
	username: string;
	week_id: string;
	points: number;
	rank: number;
	computed_at: string;
};

export type ChartDataset = {
	label: string;
	data: number[];
	borderColor: string;
	backgroundColor: string;
	tension: number;
	fill: boolean;
	borderWidth: number;
	pointRadius: number;
	pointHoverRadius: number;
};

// A user participating in a season (returned from GET /seasons/:id/participants)
export type Participant = {
	user_id: string;
	username: string | null;
	avatar_url: string;  // always returns full URL (default avatar if none set)
	joined_at: string;
};
