package token

import (
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

const (
	testJwtSecret = "test-secret-key-for-testing"
	testIssuer    = "projectmanagementapi"
	wrongSecret   = "wrong-secret"
)

func TestTokenProvider_Generate(t *testing.T) {
	cfg := &config.Config{
		JwtSecret: testJwtSecret,
	}
	provider := NewTokenProvider(cfg)

	tests := []struct {
		name      string
		sub       string
		exp       time.Time
		claims    map[string]string
		shouldErr bool
	}{
		{
			name:      "valid token generation",
			sub:       "user-123",
			exp:       time.Now().Add(time.Hour),
			claims:    map[string]string{"role": "user"},
			shouldErr: false,
		},
		{
			name:      "token with no additional claims",
			sub:       "user-456",
			exp:       time.Now().Add(time.Hour),
			claims:    nil,
			shouldErr: false,
		},
		{
			name:      "token with empty claims",
			sub:       "user-789",
			exp:       time.Now().Add(time.Hour),
			claims:    map[string]string{},
			shouldErr: false,
		},
		{
			name:      "token with multiple claims",
			sub:       "admin-123",
			exp:       time.Now().Add(time.Hour),
			claims:    map[string]string{"role": "admin", "scope": "all"},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := provider.Generate(tt.sub, tt.exp, tt.claims)

			if tt.shouldErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
					return []byte(testJwtSecret), nil
				})
				if tt.exp.Before(time.Now()) {
					assert.Error(t, err)
					assert.NotNil(t, parsedToken)
					assert.False(t, parsedToken.Valid)
					return
				} else {
					assert.NoError(t, err)
				}

				claims, ok := parsedToken.Claims.(jwt.MapClaims)
				assert.True(t, ok)

				assert.Equal(t, tt.sub, claims["sub"])
				assert.Equal(t, testIssuer, claims["iss"])
				assert.Equal(t, float64(tt.exp.Unix()), claims["exp"])
				assert.NotNil(t, claims["iat"])

				for key, value := range tt.claims {
					assert.Equal(t, value, claims[key])
				}
			}
		})
	}
}

func TestTokenProvider_Verify(t *testing.T) {
	cfg := &config.Config{
		JwtSecret: testJwtSecret,
	}
	provider := NewTokenProvider(cfg)

	validToken, _ := provider.Generate("user-123", time.Now().Add(time.Hour), map[string]string{"role": "user"})

	wrongSecretProvider := NewTokenProvider(&config.Config{JwtSecret: wrongSecret})
	wrongSecretToken, _ := wrongSecretProvider.Generate("user-456", time.Now().Add(time.Hour), nil)

	expiredToken, _ := provider.Generate("user-expired", time.Now().Add(-time.Hour), nil)

	tests := []struct {
		name      string
		token     string
		shouldErr bool
		checkSub  string
	}{
		{
			name:      "valid token",
			token:     validToken,
			shouldErr: false,
			checkSub:  "user-123",
		},
		{
			name:      "token with wrong secret",
			token:     wrongSecretToken,
			shouldErr: true,
		},
		{
			name:      "invalid token format",
			token:     "invalid.token.format",
			shouldErr: true,
		},
		{
			name:      "empty token",
			token:     "",
			shouldErr: true,
		},
		{
			name:      "malformed token",
			token:     "not-a-jwt-token",
			shouldErr: true,
		},
		{
			name:      "expired token",
			token:     expiredToken,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := provider.Verify(tt.token)

			if tt.shouldErr {
				assert.Error(t, err)
				assert.Nil(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, token)
				assert.True(t, token.Valid)

				if tt.checkSub != "" {
					claims, ok := token.Claims.(jwt.MapClaims)
					assert.True(t, ok)
					assert.Equal(t, tt.checkSub, claims["sub"])
				}
			}
		})
	}
}

func TestTokenProvider_RoundTrip(t *testing.T) {
	cfg := &config.Config{
		JwtSecret: testJwtSecret,
	}
	provider := NewTokenProvider(cfg)

	testCases := []struct {
		sub    string
		exp    time.Time
		claims map[string]string
	}{
		{
			sub:    "user-1",
			exp:    time.Now().Add(time.Hour),
			claims: map[string]string{"role": "user", "department": "engineering"},
		},
		{
			sub:    "admin-1",
			exp:    time.Now().Add(2 * time.Hour),
			claims: map[string]string{"role": "admin"},
		},
		{
			sub:    "guest-1",
			exp:    time.Now().Add(30 * time.Minute),
			claims: nil,
		},
	}

	for _, tc := range testCases {
		t.Run("round trip for "+tc.sub, func(t *testing.T) {
			tokenString, err := provider.Generate(tc.sub, tc.exp, tc.claims)
			assert.NoError(t, err)
			assert.NotEmpty(t, tokenString)

			token, err := provider.Verify(tokenString)
			assert.NoError(t, err)
			assert.NotNil(t, token)
			assert.True(t, token.Valid)

			claims, ok := token.Claims.(jwt.MapClaims)
			assert.True(t, ok)

			assert.Equal(t, tc.sub, claims["sub"])
			assert.Equal(t, testIssuer, claims["iss"])
			assert.Equal(t, float64(tc.exp.Unix()), claims["exp"])

			for key, value := range tc.claims {
				assert.Equal(t, value, claims[key])
			}
		})
	}
}
