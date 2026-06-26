package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"
	authentication "unicard-go/backend/internal/auth"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
)

type contextKey string

const UserClaimsKey contextKey = "user_claims"

// RequireAuth is a middleware that checks for a valid JWT in the HttpOnly cookie
// and ensures the user has one of the allowed roles.
func RequireAuth(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			
			// Helper to handle unauthorized/forbidden gracefully depending on endpoint type
			handleAuthError := func(w http.ResponseWriter, r *http.Request, statusCode int, msg string) {
				// Check if it's an API request (starts with /v1/ or /api/)
				if strings.HasPrefix(r.URL.Path, "/v1/") || strings.HasPrefix(r.URL.Path, "/api/") {
					jsonwrite.WriteJSON(w, statusCode, jsonwrite.APIResponse{
						Success: false,
						Message: msg,
					})
				} else {
					// It's a view, redirect to login
					http.Redirect(w, r, "/login", http.StatusSeeOther)
				}
			}

			var claims *authentication.JWTClaims
			var tokenValid bool

			// 1. Try to extract and validate Access Token
			cookie, err := r.Cookie("jwt")
			if err == nil {
				parsedClaims, err := authentication.ValidateJWT(cookie.Value)
				if err == nil && parsedClaims.Subject == "access" {
					claims = parsedClaims
					tokenValid = true
				}
			}

			// 2. If Access Token is missing or invalid, try Silent Refresh
			if !tokenValid {
				refreshCookie, err := r.Cookie("refresh_token")
				if err == nil {
					refreshClaims, err := authentication.ValidateJWT(refreshCookie.Value)
					if err == nil && refreshClaims.Subject == "refresh" {
						// Valid refresh token! Generate new tokens.
						newAccess, newRefresh, err := authentication.GenerateTokens(refreshClaims.UserID, refreshClaims.Role)
						if err == nil {
							// Set new cookies
							http.SetCookie(w, &http.Cookie{
								Name:     "jwt",
								Value:    newAccess,
								Expires:  time.Now().Add(15 * time.Minute),
								HttpOnly: true,
								Secure:   true,
								SameSite: http.SameSiteStrictMode,
								Path:     "/",
							})
							http.SetCookie(w, &http.Cookie{
								Name:     "refresh_token",
								Value:    newRefresh,
								Expires:  time.Now().Add(7 * 24 * time.Hour),
								HttpOnly: true,
								Secure:   true,
								SameSite: http.SameSiteStrictMode,
								Path:     "/",
							})
							
							// Treat request as valid using the refresh claims
							claims = refreshClaims
							tokenValid = true
						}
					}
				}
			}

			if !tokenValid {
				handleAuthError(w, r, http.StatusUnauthorized, "Unauthorized: Missing or invalid authentication token")
				return
			}

			// 3. Check role-based access control (RBAC)
			roleAllowed := false
			if len(allowedRoles) == 0 {
				// If no specific roles are required, any valid token is sufficient
				roleAllowed = true
			} else {
				for _, role := range allowedRoles {
					if claims.Role == role {
						roleAllowed = true
						break
					}
				}
			}

			if !roleAllowed {
				handleAuthError(w, r, http.StatusForbidden, "Forbidden: Insufficient privileges")
				return
			}

			// Pass the claims to the next handler via request context
			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
