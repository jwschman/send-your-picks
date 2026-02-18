// admin handlers live here
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// just returns all user accounts and their details
func GetAllAccounts(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Define a response struct that includes auth.users fields
		type UserWithAuth struct {
			ID           string  `json:"id" db:"id"`
			Username     *string `json:"username" db:"username"`
			Tagline      *string `json:"tagline" db:"tagline"`
			Role         string  `json:"role" db:"role"`
			AvatarURL    *string `json:"avatar_url" db:"avatar_url"`
			Email        string  `json:"email" db:"email"`                     // From auth.users
			LastSignInAt *string `json:"last_sign_in_at" db:"last_sign_in_at"` // From auth.users
			CreatedAt    string  `json:"created_at" db:"created_at"`
			UpdatedAt    string  `json:"updated_at" db:"updated_at"`
		}

		var users []UserWithAuth

		// join in data from auth.users to get email and lastsignin
		query := `
            SELECT
                p.id,
                p.username,
                p.tagline,
                p.role,
                p.avatar_url,
                p.created_at,
                p.updated_at,
                u.email,
                u.last_sign_in_at
            FROM public.profiles p
            INNER JOIN auth.users u ON p.id = u.id	
            ORDER BY p.created_at;
        `

		err := db.Select(&users, query)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if users == nil {
			users = []UserWithAuth{}
		}

		c.JSON(http.StatusOK, gin.H{"users": users})
	}
}
