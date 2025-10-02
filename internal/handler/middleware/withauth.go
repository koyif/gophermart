package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/koyif/gophermart/internal/config"
	"github.com/koyif/gophermart/pkg/logger"
	"net/http"
	"strings"
)

func WithAuth(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, ignore := range cfg.AuthDisabledURLs {
				if strings.HasSuffix(r.RequestURI, ignore) {
					next.ServeHTTP(w, r)
					return
				}
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				logger.Log.Warn("unauthorized request", logger.String("url", r.RequestURI))
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			var claims jwt.StandardClaims
			_, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(cfg.PrivateKey), nil
			})
			if err != nil {
				logger.Log.Warn("unauthorized request", logger.String("url", r.RequestURI), logger.Error(err))
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			r.Header.Set("User-ID", claims.Subject)

			next.ServeHTTP(w, r)
		})
	}
}
