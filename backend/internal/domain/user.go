package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	AccessTokenTTL         = 30 * time.Minute
	RefreshTokenTTL        = 3 * time.Hour
	RefreshTokenByteLength = 48
)

type User struct {
	Id        uuid.UUID `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	Email     string    `json:"email,omitempty"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type RefreshToken struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	Active    bool
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

func NewRefreshToken(userId uuid.UUID, token string, now time.Time) RefreshToken {
	return RefreshToken{
		UserId:    userId,
		Active:    true,
		Token:     token,
		CreatedAt: now,
		ExpiresAt: now.Add(RefreshTokenTTL),
	}
}

func (rt RefreshToken) IsExpiredAt(now time.Time) bool {
	return !now.Before(rt.ExpiresAt)
}

func (rt RefreshToken) IsUsableAt(now time.Time) bool {
	return rt.Active && !rt.IsExpiredAt(now)
}

func (rt *RefreshToken) Deactivate() {
	rt.Active = false
}
