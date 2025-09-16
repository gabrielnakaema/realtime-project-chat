package token

import (
	"time"

	"github.com/gabrielnakaema/project-chat/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type TokenProvider struct {
	config *config.Config
}

func NewTokenProvider(config *config.Config) *TokenProvider {
	return &TokenProvider{
		config: config,
	}
}

func (tp *TokenProvider) Generate(sub string, exp time.Time, claims map[string]string) (string, error) {
	mapClaims := jwt.MapClaims{
		"iss": "projectmanagementapi",
		"sub": sub,
		"exp": exp.Unix(),
		"iat": time.Now().Unix(),
	}

	if len(claims) > 0 {
		for k, v := range claims {
			mapClaims[k] = v
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)

	secret := []byte(tp.config.JwtSecret)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (tp *TokenProvider) Verify(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return []byte(tp.config.JwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}
