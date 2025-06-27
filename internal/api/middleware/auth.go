package middleware

import (
	"context"
	"github.com/chestorix/gophermart/internal/interfaces"
	"net/http"
	"strings"
)

func Auth(authService interfaces.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if strings.HasPrefix(token, "Bearer") {
				token = strings.TrimPrefix(token, "Bearer")
			}

			login, err := authService.ValidateToken(token)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			user, err := authService.GetUserByLogin(r.Context(), login)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "userID", user.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
