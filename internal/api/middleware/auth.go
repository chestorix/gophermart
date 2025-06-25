package middleware

import (
	"context"
	"fmt"
	"github.com/chestorix/gophermart/internal/interfaces"
	"net/http"
)

func Auth(authService interfaces.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			fmt.Println(token)
			if token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			login, err := authService.ValidateToken(token)
			fmt.Println(login)
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
