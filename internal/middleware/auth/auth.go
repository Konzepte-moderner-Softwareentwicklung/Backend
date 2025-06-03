package auth

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/jwt"
)

type AuthMiddleware struct {
	// Add fields here
	decoder jwt.Decodable
}

func NewAuthMiddleware(secret []byte) *AuthMiddleware {
	return &AuthMiddleware{
		decoder: jwt.NewDecoder(secret),
	}
}

func (m *AuthMiddleware) EnsureJWT(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		os.Stdout.WriteString(fmt.Sprintln(r.Header))
		if token == "" {
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		userID, err := m.decoder.DecodeUUID(token)
		if err != nil {
			switch err {
			case jwt.ERR_INVALID_TOKEN:
				http.Error(w, "Invalid token", http.StatusUnauthorized)
			case jwt.ERR_INVALID_CLAIMS:
				http.Error(w, "Token is expired", http.StatusUnauthorized)
			default:
				http.Error(w, "Unknown error", http.StatusInternalServerError)
			}
			return
		}
		r.Header.Add("UserId", userID.String())
		next.ServeHTTP(w, r)
	})
}
