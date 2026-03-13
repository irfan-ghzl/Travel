package gapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	db "github.com/irfan-ghzl/pintour/internal/db/sqlc"
	"github.com/irfan-ghzl/pintour/internal/middleware"
	"github.com/irfan-ghzl/pintour/internal/util"
	pb "github.com/irfan-ghzl/pintour/pb/pintour/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Register creates a new user account
func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email, password, and name are required")
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot hash password: %v", err)
	}

	user, err := s.store.CreateUser(ctx, db.CreateUserParams{
		Email:        req.Email,
		Name:         req.Name,
		Phone:        req.Phone,
		PasswordHash: hashedPassword,
		GoogleID:     "",
		AvatarUrl:    "",
		Role:         "user",
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create user: %v", err)
	}

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(user.ID, user.Email, user.Role, s.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create access token: %v", err)
	}

	return &pb.RegisterResponse{
		User:                convertUser(user),
		AccessToken:         accessToken,
		AccessTokenExpiresAt: timestamppb.New(accessPayload.ExpiredAt),
	}, nil
}

// Login authenticates a user with email and password
func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email and password are required")
	}

	user, err := s.store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot get user: %v", err)
	}

	err = util.CheckPassword(req.Password, user.PasswordHash)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "incorrect password")
	}

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(user.ID, user.Email, user.Role, s.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create access token: %v", err)
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(user.ID, user.Email, user.Role, s.config.RefreshTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create refresh token: %v", err)
	}

	// Extract metadata for session
	var userAgent, clientIP string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ua := md.Get("user-agent"); len(ua) > 0 {
			userAgent = ua[0]
		}
		if ip := md.Get("x-forwarded-for"); len(ip) > 0 {
			clientIP = ip[0]
		}
	}

	session, err := s.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		ClientIp:     clientIP,
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create session: %v", err)
	}

	return &pb.LoginResponse{
		User:                  convertUser(user),
		SessionId:             session.ID.String(),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  timestamppb.New(accessPayload.ExpiredAt),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiredAt),
	}, nil
}

// googleUserInfo holds user info from Google OAuth
type googleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// GoogleLogin authenticates a user via Google OAuth2
func (s *Server) GoogleLogin(ctx context.Context, req *pb.GoogleLoginRequest) (*pb.GoogleLoginResponse, error) {
	if req.Code == "" {
		return nil, status.Errorf(codes.InvalidArgument, "authorization code is required")
	}

	oauthToken, err := s.oauthConfig.Exchange(ctx, req.Code)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to exchange oauth code: %v", err)
	}

	client := s.oauthConfig.Client(ctx, oauthToken)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, status.Errorf(codes.Internal, "google userinfo returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read user info: %v", err)
	}

	var userInfo googleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse user info: %v", err)
	}

	// Check if user exists by Google ID
	user, err := s.store.GetUserByGoogleID(ctx, userInfo.ID)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, status.Errorf(codes.Internal, "cannot get user: %v", err)
		}
		// Try by email
		user, err = s.store.GetUserByEmail(ctx, userInfo.Email)
		if err != nil {
			if err != sql.ErrNoRows {
				return nil, status.Errorf(codes.Internal, "cannot get user: %v", err)
			}
			// Create new user
			user, err = s.store.CreateUser(ctx, db.CreateUserParams{
				Email:        userInfo.Email,
				Name:         userInfo.Name,
				Phone:        "",
				PasswordHash: "",
				GoogleID:     userInfo.ID,
				AvatarUrl:    userInfo.Picture,
				Role:         "user",
			})
			if err != nil {
				return nil, status.Errorf(codes.Internal, "cannot create user: %v", err)
			}
		}
	}

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(user.ID, user.Email, user.Role, s.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create access token: %v", err)
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(user.ID, user.Email, user.Role, s.config.RefreshTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create refresh token: %v", err)
	}

	sessionID, _ := uuid.NewRandom()
	session, err := s.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           sessionID,
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    "",
		ClientIp:     "",
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create session: %v", err)
	}

	return &pb.GoogleLoginResponse{
		User:                  convertUser(user),
		SessionId:             session.ID.String(),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  timestamppb.New(accessPayload.ExpiredAt),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiredAt),
	}, nil
}

// GetProfile returns the authenticated user's profile
func (s *Server) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.User, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}

	user, err := s.store.GetUserByID(ctx, payload.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot get user: %v", err)
	}

	return convertUser(user), nil
}

// UpdateProfile updates the authenticated user's profile
func (s *Server) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.User, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}

	var name, phone, avatarURL sql.NullString
	if req.Name != "" {
		name = sql.NullString{String: req.Name, Valid: true}
	}
	if req.Phone != "" {
		phone = sql.NullString{String: req.Phone, Valid: true}
	}
	if req.AvatarUrl != "" {
		avatarURL = sql.NullString{String: req.AvatarUrl, Valid: true}
	}

	user, err := s.store.UpdateUser(ctx, db.UpdateUserParams{
		ID:        payload.UserID,
		Name:      name,
		Phone:     phone,
		AvatarUrl: avatarURL,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot update user: %v", err)
	}

	return convertUser(user), nil
}

// convertUser converts a db.User to a pb.User
func convertUser(user db.User) *pb.User {
	role := pb.UserRole_USER_ROLE_USER
	if user.Role == "admin" {
		role = pb.UserRole_USER_ROLE_ADMIN
	}

	return &pb.User{
		Id:         user.ID,
		Email:      user.Email,
		Name:       user.Name,
		Phone:      user.Phone,
		AvatarUrl:  user.AvatarUrl,
		Role:       role,
		IsVerified: user.IsVerified,
		CreatedAt:  timestamppb.New(user.CreatedAt),
		UpdatedAt:  timestamppb.New(user.UpdatedAt),
	}
}

// ensure time is used
var _ = fmt.Sprintf
var _ = time.Now
