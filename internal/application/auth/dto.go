package auth

import (
	"time"

	"github.com/irfan-ghzl/pintour/internal/domain/entity"
)

type RegisterInput struct {
	Email    string
	Password string
	Name     string
	Phone    string
}

type RegisterOutput struct {
	User                 *entity.User
	AccessToken          string
	AccessTokenExpiresAt time.Time
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	User                  *entity.User
	SessionID             string
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}

type GoogleLoginInput struct {
	Code      string
	UserAgent string
	ClientIP  string
}

type UpdateProfileInput struct {
	UserID    int64
	Name      *string
	Phone     *string
	AvatarURL *string
}
