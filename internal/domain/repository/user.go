package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/irfan-ghzl/pintour/internal/domain/entity"
)

type CreateUserParams struct {
	Email        string
	Name         string
	Phone        string
	PasswordHash string
	GoogleID     string
	AvatarURL    string
	Role         entity.UserRole
}

type UpdateUserParams struct {
	ID        int64
	Name      *string
	Phone     *string
	AvatarURL *string
}

type UserRepository interface {
	Create(ctx context.Context, params CreateUserParams) (*entity.User, error)
	GetByID(ctx context.Context, id int64) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByGoogleID(ctx context.Context, googleID string) (*entity.User, error)
	Update(ctx context.Context, params UpdateUserParams) (*entity.User, error)
	List(ctx context.Context, limit, offset int32) ([]entity.User, error)
	Count(ctx context.Context) (int64, error)
}

type CreateSessionParams struct {
	ID           uuid.UUID
	UserID       int64
	RefreshToken string
	UserAgent    string
	ClientIP     string
	IsBlocked    bool
	ExpiresAt    interface{}
}

type SessionRepository interface {
	Create(ctx context.Context, params CreateSessionParams) (interface{}, error)
	GetByID(ctx context.Context, id uuid.UUID) (interface{}, error)
}
