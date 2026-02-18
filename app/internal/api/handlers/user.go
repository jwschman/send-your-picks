package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"

	"github.com/gin-gonic/gin"
	"pawked.com/sendyourpicks/internal/api/middleware"
	"pawked.com/sendyourpicks/internal/models"
)

// buildAvatarURL takes an avatar path from the database and returns the full URL.
// If avatarPath is nil, returns the default avatar URL.
// I might actually ditch the default avatar because having initials may be better... we'll see what people say
func buildAvatarURL(avatarPath *string) string {
	supabaseURL := os.Getenv("SUPABASE_STORAGE_URL")
	siteAddress := os.Getenv("SITE_ADDRESS")

	if avatarPath == nil {
		return fmt.Sprintf("%s/images/avatars/default-avatar.jpg", siteAddress)
	}
	return fmt.Sprintf("%s/storage/v1/object/public/avatars/%s", supabaseURL, *avatarPath)
}

// WhoAmI returns authenticated user identity information.
// It requires a valid JWT (validated by middleware) and adds username and avatar from the profiles table.
func WhoAmI(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := middleware.GetClaims(c)

		var userInfo struct {
			UserName  string  `db:"username" json:"username"`
			AvatarURL *string `db:"avatar_url" json:"avatar_url"`
		}

		err := db.Get(
			&userInfo,
			`SELECT username, avatar_url FROM public.profiles WHERE id = $1`,
			claims.Sub,
		)

		if err != nil && err != sql.ErrNoRows {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load profile"})
			return
		}

		avatarURL := buildAvatarURL(userInfo.AvatarURL)

		c.JSON(http.StatusOK, gin.H{
			"id":         claims.Sub,
			"email":      claims.Email,
			"role":       claims.UserRole,
			"username":   userInfo.UserName,
			"avatar_url": avatarURL,
		})
	}
}

// GetAccount fetches the full user profile from the database.
func GetMyAccount(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := middleware.GetClaims(c)

		userID := claims.Sub
		userEmail := claims.Email

		// Use the User model
		var user models.User

		// build the SQL query.  $1 is going to be userID
		query := `SELECT * FROM public.profiles WHERE id = $1`
		// sqlx.Get automatically maps columns to struct fields using db tags -- this is so much easier than before
		err := db.Get(&user, query, userID)

		if err != nil {
			// if we can't find the user
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
				return
			}
			// general error
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// transform avatar path to full URL

		// Return the profile with email from JWT
		c.JSON(http.StatusOK, gin.H{
			"id":         user.ID,
			"email":      userEmail, // From JWT
			"username":   user.Username,
			"tagline":    user.Tagline,
			"role":       user.Role,
			"avatar_url": user.AvatarURL,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		})
	}
}

// UpdateAccount allows users to update their own profile information.
func UpdateAccount(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := middleware.GetClaims(c)

		userID := claims.Sub
		userEmail := claims.Email

		// Parse request body
		var req struct {
			Username  *string `json:"username"`
			Tagline   *string `json:"tagline"`
			AvatarURL *string `json:"avatar_url"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Use transaction for explicit control
		tx, err := db.Beginx()
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer tx.Rollback()

		query := `
			UPDATE public.profiles
			SET
				username = $1,
				tagline = $2,
				avatar_url = $3,
				updated_at = NOW()
			WHERE id = $4
			RETURNING *
		`

		var user models.User
		err = tx.QueryRowx(
			query,
			req.Username,
			req.Tagline,
			req.AvatarURL,
			userID,
		).StructScan(&user)

		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}

		if err := tx.Commit(); err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit update"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":         user.ID,
			"email":      userEmail,
			"username":   user.Username,
			"tagline":    user.Tagline,
			"role":       user.Role,
			"avatar_url": user.AvatarURL,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		})
	}
}

// GetAllUsers fetches a list of users for public viewing.
func GetAllUsers(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		var users []models.PublicProfile
		query := `
            SELECT
                id,
                username,
                tagline,
                role,
                avatar_url
            FROM public.profiles 
            ORDER BY username;
        `

		err := db.Select(&users, query)
		if err != nil {
			// Log the actual error for debugging
			c.Error(err) // log to gin's error logger
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Return empty array instead of null if no users
		if users == nil {
			users = []models.PublicProfile{}
		}

		// Transform avatar paths to full URLs
		for i := range users {
			fullURL := buildAvatarURL(users[i].AvatarURL)
			users[i].AvatarURL = &fullURL
		}

		c.JSON(http.StatusOK, gin.H{"users": users})
	}
}

// GetUserAccount fetches a public user profile from the database.
func GetUser(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		userID := c.Param("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user ID"})
			return
		}

		var profile models.PublicProfile
		query := `
            SELECT id, username, tagline, role, avatar_url 
            FROM public.profiles 
            WHERE id = $1
        `

		err := db.Get(&profile, query, userID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Transform avatar path to full URL
		fullURL := buildAvatarURL(profile.AvatarURL)
		profile.AvatarURL = &fullURL

		c.JSON(http.StatusOK, gin.H{"user": profile})
	}
}
