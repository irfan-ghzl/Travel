package persistence

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	db "github.com/irfan-ghzl/pintour/internal/db/sqlc"
	"github.com/irfan-ghzl/pintour/internal/domain/entity"
	"github.com/irfan-ghzl/pintour/internal/domain/repository"
)

type userRepository struct {
	q db.Querier
}

func NewUserRepository(q db.Querier) repository.UserRepository {
	return &userRepository{q: q}
}

func (r *userRepository) Create(ctx context.Context, params repository.CreateUserParams) (*entity.User, error) {
	u, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Email:        params.Email,
		Name:         params.Name,
		Phone:        params.Phone,
		PasswordHash: params.PasswordHash,
		GoogleID:     params.GoogleID,
		AvatarUrl:    params.AvatarURL,
		Role:         string(params.Role),
	})
	if err != nil {
		return nil, err
	}
	return toEntityUser(u), nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	u, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toEntityUser(u), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	u, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return toEntityUser(u), nil
}

func (r *userRepository) GetByGoogleID(ctx context.Context, googleID string) (*entity.User, error) {
	u, err := r.q.GetUserByGoogleID(ctx, googleID)
	if err != nil {
		return nil, err
	}
	return toEntityUser(u), nil
}

func (r *userRepository) Update(ctx context.Context, params repository.UpdateUserParams) (*entity.User, error) {
	var name, phone, avatarURL sql.NullString
	if params.Name != nil {
		name = sql.NullString{String: *params.Name, Valid: true}
	}
	if params.Phone != nil {
		phone = sql.NullString{String: *params.Phone, Valid: true}
	}
	if params.AvatarURL != nil {
		avatarURL = sql.NullString{String: *params.AvatarURL, Valid: true}
	}

	u, err := r.q.UpdateUser(ctx, db.UpdateUserParams{
		ID:        params.ID,
		Name:      name,
		Phone:     phone,
		AvatarUrl: avatarURL,
	})
	if err != nil {
		return nil, err
	}
	return toEntityUser(u), nil
}

func (r *userRepository) List(ctx context.Context, limit, offset int32) ([]entity.User, error) {
	users, err := r.q.ListUsers(ctx, db.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	var result []entity.User
	for _, u := range users {
		result = append(result, *toEntityUser(u))
	}
	return result, nil
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	return r.q.CountUsers(ctx)
}

type sessionRepository struct {
	q db.Querier
}

func NewSessionRepository(q db.Querier) repository.SessionRepository {
	return &sessionRepository{q: q}
}

func (r *sessionRepository) Create(ctx context.Context, params repository.CreateSessionParams) (interface{}, error) {
	return nil, nil
}

func (r *sessionRepository) GetByID(ctx context.Context, id uuid.UUID) (interface{}, error) {
	return r.q.GetSession(ctx, id)
}

func toEntityUser(u db.User) *entity.User {
	role := entity.UserRoleUser
	if u.Role == "admin" {
		role = entity.UserRoleAdmin
	}
	return &entity.User{
		ID:           u.ID,
		Email:        u.Email,
		Name:         u.Name,
		Phone:        u.Phone,
		PasswordHash: u.PasswordHash,
		GoogleID:     u.GoogleID,
		AvatarURL:    u.AvatarUrl,
		Role:         role,
		IsVerified:   u.IsVerified,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}
