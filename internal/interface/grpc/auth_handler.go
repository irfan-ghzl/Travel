package grpc

import (
	"context"

	"github.com/irfan-ghzl/pintour/internal/application/auth"
	"github.com/irfan-ghzl/pintour/internal/domain/entity"
	"github.com/irfan-ghzl/pintour/internal/interface/middleware"
	pb "github.com/irfan-ghzl/pintour/pb/pintour/v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	out, err := s.authService.Register(ctx, auth.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
		Phone:    req.Phone,
	})
	if err != nil {
		return nil, err
	}
	return &pb.RegisterResponse{
		User:                 convertUser(out.User),
		AccessToken:          out.AccessToken,
		AccessTokenExpiresAt: timestamppb.New(out.AccessTokenExpiresAt),
	}, nil
}

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	var userAgent, clientIP string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ua := md.Get("user-agent"); len(ua) > 0 {
			userAgent = ua[0]
		}
		if ip := md.Get("x-forwarded-for"); len(ip) > 0 {
			clientIP = ip[0]
		}
	}
	_ = userAgent
	_ = clientIP

	out, err := s.authService.Login(ctx, auth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &pb.LoginResponse{
		User:                  convertUser(out.User),
		SessionId:             out.SessionID,
		AccessToken:           out.AccessToken,
		AccessTokenExpiresAt:  timestamppb.New(out.AccessTokenExpiresAt),
		RefreshToken:          out.RefreshToken,
		RefreshTokenExpiresAt: timestamppb.New(out.RefreshTokenExpiresAt),
	}, nil
}

func (s *Server) GoogleLogin(ctx context.Context, req *pb.GoogleLoginRequest) (*pb.GoogleLoginResponse, error) {
	out, err := s.authService.GoogleLogin(ctx, auth.GoogleLoginInput{Code: req.Code})
	if err != nil {
		return nil, err
	}
	return &pb.GoogleLoginResponse{
		User:                  convertUser(out.User),
		SessionId:             out.SessionID,
		AccessToken:           out.AccessToken,
		AccessTokenExpiresAt:  timestamppb.New(out.AccessTokenExpiresAt),
		RefreshToken:          out.RefreshToken,
		RefreshTokenExpiresAt: timestamppb.New(out.RefreshTokenExpiresAt),
	}, nil
}

func (s *Server) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.User, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}
	user, err := s.authService.GetProfile(ctx, payload.UserID)
	if err != nil {
		return nil, err
	}
	return convertUser(user), nil
}

func (s *Server) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.User, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}

	input := auth.UpdateProfileInput{UserID: payload.UserID}
	if req.Name != "" {
		input.Name = &req.Name
	}
	if req.Phone != "" {
		input.Phone = &req.Phone
	}
	if req.AvatarUrl != "" {
		input.AvatarURL = &req.AvatarUrl
	}

	user, err := s.authService.UpdateProfile(ctx, input)
	if err != nil {
		return nil, err
	}
	return convertUser(user), nil
}

func convertUser(u *entity.User) *pb.User {
	role := pb.UserRole_USER_ROLE_USER
	if u.Role == entity.UserRoleAdmin {
		role = pb.UserRole_USER_ROLE_ADMIN
	}
	return &pb.User{
		Id:         u.ID,
		Email:      u.Email,
		Name:       u.Name,
		Phone:      u.Phone,
		AvatarUrl:  u.AvatarURL,
		Role:       role,
		IsVerified: u.IsVerified,
		CreatedAt:  timestamppb.New(u.CreatedAt),
		UpdatedAt:  timestamppb.New(u.UpdatedAt),
	}
}
