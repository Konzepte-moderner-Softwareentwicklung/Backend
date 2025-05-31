package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Encoder struct {
	key []byte
}

func NewEncoder(key []byte) *Encoder {
	return &Encoder{
		key: key,
	}
}

func (e *Encoder) EncodeUUID(id uuid.UUID, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"uuid": id.String(),
		"exp":  time.Now().Add(ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(e.key)
}
