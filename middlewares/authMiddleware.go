package middleware

import (
	"context"
	"net/http"
	"strings"

	helper "github.com/02priyeshraj/Hotel_Management_Backend/helper"
)

// Context keys to store user information
type contextKey string

const (
	EmailKey     contextKey = "email"
	FirstNameKey contextKey = "first_name"
	LastNameKey  contextKey = "last_name"
	UidKey       contextKey = "uid"
)

// Authentication middleware for Gorilla Mux
func Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientToken := r.Header.Get("Authorization")
		if clientToken == "" {
			http.Error(w, "No Authorization header provided", http.StatusUnauthorized)
			return
		}

		// Token format should be "Bearer <token>"
		tokenParts := strings.Split(clientToken, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization format", http.StatusUnauthorized)
			return
		}

		tokenString := tokenParts[1]
		claims, err := helper.ValidateToken(tokenString)
		if err != "" {
			http.Error(w, err, http.StatusUnauthorized)
			return
		}

		// Store user details in the request context
		ctx := context.WithValue(r.Context(), EmailKey, claims.Email)
		ctx = context.WithValue(ctx, FirstNameKey, claims.FirstName)
		ctx = context.WithValue(ctx, LastNameKey, claims.LastName)
		ctx = context.WithValue(ctx, UidKey, claims.Uid)

		// Pass modified request with context to the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext retrieves user data from the request context
func GetUserFromContext(r *http.Request) (email, firstName, lastName, uid string) {
	email, _ = r.Context().Value(EmailKey).(string)
	firstName, _ = r.Context().Value(FirstNameKey).(string)
	lastName, _ = r.Context().Value(LastNameKey).(string)
	uid, _ = r.Context().Value(UidKey).(string)
	return
}
