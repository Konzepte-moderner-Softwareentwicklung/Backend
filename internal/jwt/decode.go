package jwt

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrInvalidClaims = errors.New("invalid claims")
	ErrInvalidUUID   = errors.New("invalid uuid")
)

type Decodable interface {
	DecodeUUID(tokenString string) (uuid.UUID, error)
}

type Decoder struct {
	key []byte
}

func NewDecoder(key []byte) *Decoder {
	return &Decoder{key: key}
}

func (d *Decoder) DecodeUUID(tokenString string) (uuid.UUID, error) {
	// Token parsen
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Sicherstellen, dass der Signatur-Algorithmus stimmt
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("ungültiger Signatur-Algorithmus")
		}
		return d.key, nil
	})

	// Wenn Parsen fehlgeschlagen ist oder Token ungültig ist
	if err != nil || !token.Valid {
		return uuid.UUID{}, ErrInvalidToken
	}

	// Claims extrahieren
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.UUID{}, ErrInvalidClaims
	}

	id, ok := claims["uuid"].(string)
	if !ok {
		return uuid.UUID{}, ErrInvalidClaims
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, ErrInvalidUUID
	}

	return uid, nil
}
