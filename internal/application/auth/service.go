package auth

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/irfan-ghzl/pintour/common/config"
	"github.com/irfan-ghzl/pintour/internal/domain/entity"
	"github.com/irfan-ghzl/pintour/internal/domain/repository"
	"github.com/irfan-ghzl/pintour/internal/infrastructure/oauth"
	"github.com/irfan-ghzl/pintour/common/token"
	"github.com/irfan-ghzl/pintour/common/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	userRepo      repository.UserRepository
	tokenMaker    token.Maker
	oauthProvider oauth.GoogleProvider
	config        config.Config
}

func NewService(
	userRepo repository.UserRepository,
	tokenMaker token.Maker,
	oauthProvider oauth.GoogleProvider,
	cfg config.Config,
) *Service {
	return &Service{
		userRepo:      userRepo,
		tokenMaker:    tokenMaker,
		oauthProvider: oauthProvider,
		config:        cfg,
	}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	if input.Email == "" || input.Password == "" || input.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email, password, and name are required")
	}

	hashedPassword, err := util.HashPassword(input.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot hash password: %v", err)
	}

	user, err := s.userRepo.Create(ctx, repository.CreateUserParams{
		Email:        input.Email,
		Name:         input.Name,
		Phone:        input.Phone,
		PasswordHash: hashedPassword,
		GoogleID:     "",
		AvatarURL:    "",
		Role:         entity.UserRoleUser,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create user: %v", err)
	}

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(user.ID, user.Email, string(user.Role), s.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create token: %v", err)
	}

	return &RegisterOutput{
		User:                 user,
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	if input.Email == "" || input.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email and password are required")
	}

	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot get user: %v", err)
	}

	if err := util.CheckPassword(input.Password, user.PasswordHash); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "incorrect password")
	}

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(user.ID, user.Email, string(user.Role), s.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create access token: %v", err)
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(user.ID, user.Email, string(user.Role), s.config.RefreshTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create refresh token: %v", err)
	}

	sessionID := uuid.New()

	return &LoginOutput{
		User:                  user,
		SessionID:             sessionID.String(),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
	}, nil
}

func (s *Service) GoogleLogin(ctx context.Context, input GoogleLoginInput) (*LoginOutput, error) {
	if input.Code == "" {
		return nil, status.Errorf(codes.InvalidArgument, "authorization code is required")
	}

	userInfo, err := s.oauthProvider.GetUserInfo(ctx, input.Code)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to get google user info: %v", err)
	}

	user, err := s.userRepo.GetByGoogleID(ctx, userInfo.ID)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, status.Errorf(codes.Internal, "cannot get user: %v", err)
		}
		user, err = s.userRepo.GetByEmail(ctx, userInfo.Email)
		if err != nil {
			if err != sql.ErrNoRows {
				return nil, status.Errorf(codes.Internal, "cannot get user: %v", err)
			}
			user, err = s.userRepo.Create(ctx, repository.CreateUserParams{
				Email:        userInfo.Email,
				Name:         userInfo.Name,
				Phone:        "",
				PasswordHash: "",
				GoogleID:     userInfo.ID,
				AvatarURL:    userInfo.Picture,
				Role:         entity.UserRoleUser,
			})
			if err != nil {
				return nil, status.Errorf(codes.Internal, "cannot create user: %v", err)
			}
		}
	}

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(user.ID, user.Email, string(user.Role), s.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create access token: %v", err)
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(user.ID, user.Email, string(user.Role), s.config.RefreshTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create refresh token: %v", err)
	}

	return &LoginOutput{
		User:                  user,
		SessionID:             uuid.New().String(),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
	}, nil
}

func (s *Service) GetProfile(ctx context.Context, userID int64) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot get user: %v", err)
	}
	return user, nil
}

func (s *Service) UpdateProfile(ctx context.Context, input UpdateProfileInput) (*entity.User, error) {
	user, err := s.userRepo.Update(ctx, repository.UpdateUserParams{
		ID:        input.UserID,
		Name:      input.Name,
		Phone:     input.Phone,
		AvatarURL: input.AvatarURL,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot update user: %v", err)
	}
	return user, nil
}
