package middleware

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthError string

const (
	AuthErrorUnauthorized AuthError = "unauthorized"
	AuthErrorForbidden    AuthError = "forbidden"
	AuthErrorInvalidToken AuthError = "invalid_token"
)

type SupabaseClaims struct {
	Sub      string `json:"sub"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	UserRole string `json:"user_role"`
	jwt.RegisteredClaims
}

type AuthConfig struct {
	JWTIssuer   string
	JWKSURL     string
	keyCache    map[string]*ecdsa.PublicKey
	cacheMutex  sync.RWMutex
	cacheExpiry time.Time
}

// JWKS structures
type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

// LoadAuthConfig reads Supabase JWT config from environment variables and optionally preloads signing keys
func LoadAuthConfig() (*AuthConfig, error) {
	issuer := os.Getenv("SUPABASE_JWT_ISSUER")
	if issuer == "" {
		return nil, fmt.Errorf("SUPABASE_JWT_ISSUER is required")
	}

	jwksURL := os.Getenv("SUPABASE_JWKS_URL")
	if jwksURL == "" {
		return nil, fmt.Errorf("SUPABASE_JWKS_URL is required")
	}

	cfg := &AuthConfig{
		JWTIssuer: issuer,
		JWKSURL:   jwksURL,
		keyCache:  make(map[string]*ecdsa.PublicKey),
	}

	// Optional preload from env
	if keysJSON := os.Getenv("GOTRUE_JWT_KEYS"); keysJSON != "" {
		_ = cfg.parseJWTKeysFromEnv(keysJSON)
	}

	return cfg, nil
}

// parseJWTKeysFromEnv parses EC public keys from the GOTRUE_JWT_KEYS env var into the cache
func (cfg *AuthConfig) parseJWTKeysFromEnv(keysJSON string) error {
	var keys []JWK
	if err := json.Unmarshal([]byte(keysJSON), &keys); err != nil {
		return err
	}

	cfg.cacheMutex.Lock()
	defer cfg.cacheMutex.Unlock()

	for _, jwk := range keys {
		if jwk.Kty != "EC" || jwk.Kid == "" {
			continue
		}

		xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
		if err != nil {
			continue
		}
		yBytes, err := base64.RawURLEncoding.DecodeString(jwk.Y)
		if err != nil {
			continue
		}

		cfg.keyCache[jwk.Kid] = &ecdsa.PublicKey{
			Curve: getCurve(jwk.Crv),
			X:     new(big.Int).SetBytes(xBytes),
			Y:     new(big.Int).SetBytes(yBytes),
		}
	}

	cfg.cacheExpiry = time.Now().Add(24 * time.Hour)
	return nil
}

// fetchJWKS retrieves the JSON Web Key Set from the Supabase JWKS endpoint
func (cfg *AuthConfig) fetchJWKS() (*JWKS, error) {
	resp, err := http.Get(cfg.JWKSURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jwks endpoint returned %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}

	return &jwks, nil
}

// getPublicKey returns the cached public key for a key ID, fetching from JWKS if missing or expired
func (cfg *AuthConfig) getPublicKey(kid string) (*ecdsa.PublicKey, error) {
	cfg.cacheMutex.RLock()
	if key, ok := cfg.keyCache[kid]; ok && time.Now().Before(cfg.cacheExpiry) {
		cfg.cacheMutex.RUnlock()
		return key, nil
	}
	cfg.cacheMutex.RUnlock()

	jwks, err := cfg.fetchJWKS()
	if err != nil {
		return nil, err
	}

	for _, jwk := range jwks.Keys {
		if jwk.Kid != kid || jwk.Kty != "EC" {
			continue
		}

		x, err := base64.RawURLEncoding.DecodeString(jwk.X)
		if err != nil {
			return nil, err
		}
		y, err := base64.RawURLEncoding.DecodeString(jwk.Y)
		if err != nil {
			return nil, err
		}

		key := &ecdsa.PublicKey{
			Curve: getCurve(jwk.Crv),
			X:     new(big.Int).SetBytes(x),
			Y:     new(big.Int).SetBytes(y),
		}

		cfg.cacheMutex.Lock()
		cfg.keyCache[kid] = key
		cfg.cacheExpiry = time.Now().Add(time.Hour)
		cfg.cacheMutex.Unlock()

		return key, nil
	}

	return nil, fmt.Errorf("kid %s not found", kid)
}

// getCurve maps a JWK curve name to the corresponding Go elliptic curve
func getCurve(crv string) elliptic.Curve {
	switch crv {
	case "P-256":
		return elliptic.P256()
	case "P-384":
		return elliptic.P384()
	case "P-521":
		return elliptic.P521()
	default:
		return elliptic.P256()
	}
}

// AuthMiddleware validates the JWT from the Authorization header and stores claims in the context
func AuthMiddleware(cfg *AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   AuthErrorUnauthorized,
				"message": "authentication required",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   AuthErrorInvalidToken,
				"message": "invalid authorization header",
			})
			c.Abort()
			return
		}

		claims := &SupabaseClaims{}

		parser := jwt.NewParser(
			jwt.WithValidMethods([]string{"ES256"}),
			jwt.WithIssuer(cfg.JWTIssuer),
			jwt.WithAudience("authenticated"),
		)

		token, err := parser.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, fmt.Errorf("missing kid")
			}
			return cfg.getPublicKey(kid)
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   AuthErrorInvalidToken,
				"message": "invalid or expired token",
			})
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}

// RequireRole returns middleware that restricts access to users with one of the specified roles
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawClaims, ok := c.Get("claims")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   AuthErrorUnauthorized,
				"message": "authentication required",
			})
			c.Abort()
			return
		}

		claims, ok := rawClaims.(*SupabaseClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   AuthErrorUnauthorized,
				"message": "authentication required",
			})
			c.Abort()
			return
		}

		for _, role := range allowedRoles {
			if claims.UserRole == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":   AuthErrorForbidden,
			"message": "insufficient permissions",
		})
		c.Abort()
	}
}

// GetUserID extracts the authenticated user's ID from the context
func GetUserID(c *gin.Context) string {
	return c.MustGet("claims").(*SupabaseClaims).Sub
}

// GetClaims extracts the full JWT claims from the context
func GetClaims(c *gin.Context) *SupabaseClaims {
	return c.MustGet("claims").(*SupabaseClaims)
}
