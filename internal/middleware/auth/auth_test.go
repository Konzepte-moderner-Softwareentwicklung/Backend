package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/jwt"
	"github.com/google/uuid"
)

// injectDecoder replaces the real decoder for testing
func injectDecoder(mw *AuthMiddleware, decoder jwt.Decodable) {
	mw.decoder = decoder
}

func TestEnsureJWT(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		mockDecodeFunc func(token string) (uuid.UUID, error)
		wantStatus     int
		wantUserHeader bool
	}{
		{
			name:       "Missing Authorization Header",
			authHeader: "",
			mockDecodeFunc: func(token string) (uuid.UUID, error) {
				return uuid.Nil, nil
			},
			wantStatus:     http.StatusUnauthorized,
			wantUserHeader: false,
		},
		{
			name:       "Invalid Token",
			authHeader: "invalid-token",
			mockDecodeFunc: func(token string) (uuid.UUID, error) {
				return uuid.Nil, jwt.ERR_INVALID_TOKEN
			},
			wantStatus:     http.StatusUnauthorized,
			wantUserHeader: false,
		},
		{
			name:       "Expired Token",
			authHeader: "expired-token",
			mockDecodeFunc: func(token string) (uuid.UUID, error) {
				return uuid.Nil, jwt.ERR_INVALID_CLAIMS
			},
			wantStatus:     http.StatusUnauthorized,
			wantUserHeader: false,
		},
		{
			name:       "Valid Token",
			authHeader: "valid-token",
			mockDecodeFunc: func(token string) (uuid.UUID, error) {
				return uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), nil
			},
			wantStatus:     http.StatusOK,
			wantUserHeader: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mw := NewAuthMiddleware([]byte("secret"))
			injectDecoder(mw, &jwt.MockDecoder{DecodeFunc: tc.mockDecodeFunc})

			handlerCalled := false
			testHandler := mw.EnsureJWT(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				if tc.wantUserHeader {
					if r.Header.Get("UserId") == "" {
						t.Error("UserId header not set")
					}
				}
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rec := httptest.NewRecorder()

			testHandler.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("expected status %d, got %d", tc.wantStatus, rec.Code)
			}

			if tc.wantStatus == http.StatusOK && !handlerCalled {
				t.Error("handler should have been called on valid token")
			}
		})
	}
}
