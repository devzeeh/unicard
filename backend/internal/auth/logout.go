package authentication

import (
	"net/http"
	"time"
)

// LogoutHandler clears the authentication cookies and redirects the user to the login page.
func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear the access token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	// Clear the refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
