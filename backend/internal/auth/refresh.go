package authentication

import (
	"log"
	"net/http"
	"time"
	jsonwrite "unicard-go/backend/internal/pkg/handler"
)

// RefreshTokenHandler handles the generation of new access tokens using a valid refresh token.
func (h *Handler) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get the refresh token from the cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusUnauthorized, jsonwrite.APIResponse{
			Success: false,
			Message: "Unauthorized: Missing refresh token",
		})
		return
	}

	// 2. Validate the refresh token
	claims, err := ValidateJWT(cookie.Value)
	if err != nil || claims.Subject != "refresh" {
		log.Printf("Invalid refresh token attempt: %v", err)
		jsonwrite.WriteJSON(w, http.StatusUnauthorized, jsonwrite.APIResponse{
			Success: false,
			Message: "Unauthorized: Invalid or expired refresh token",
		})
		return
	}

	// 3. Generate new tokens
	accessToken, newRefreshToken, err := GenerateTokens(claims.UserID, claims.Role)
	if err != nil {
		log.Printf("Error generating tokens during refresh: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	// 4. Set the new Access Token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    accessToken,
		Expires:  time.Now().Add(15 * time.Minute),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	// 5. Set the new Refresh Token cookie (sliding expiration)
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	// 6. Return success response
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Token refreshed successfully",
	})
}
