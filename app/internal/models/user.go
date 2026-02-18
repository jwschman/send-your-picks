package models

import "time"

// User represents a user profile in the database
// corresponds to the public.profiles table.
type User struct {
	ID        string    `json:"id" db:"id"`                 // user's UUID - matches auth.users(id)
	Username  *string   `json:"username" db:"username"`     // display name - defaults to email on signup - nullable in database so use pointer
	Tagline   *string   `json:"tagline" db:"tagline"`       // optional title - also nullable
	Role      string    `json:"role" db:"role"`             // user's permission level - maps to user_role enum
	AvatarURL *string   `json:"avatar_url" db:"avatar_url"` // nullable
	CreatedAt time.Time `json:"created_at" db:"created_at"` // when the profile was created
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // last modified timestamp
}

// For viewing by by other users
type PublicProfile struct {
	ID        string  `json:"id" db:"id"`
	Username  *string `json:"username" db:"username"`
	Tagline   *string `json:"tagline" db:"tagline"`
	Role      string  `json:"role" db:"role"`
	AvatarURL *string `json:"avatar_url" db:"avatar_url"`
}
