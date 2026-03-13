package gapi

import (
	"fmt"

	"github.com/irfan-ghzl/pintour/internal/config"
	db "github.com/irfan-ghzl/pintour/internal/db/sqlc"
	"github.com/irfan-ghzl/pintour/internal/token"
	pb "github.com/irfan-ghzl/pintour/pb/pintour/v1"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Server serves gRPC requests for the PINTOUR service
type Server struct {
	pb.UnimplementedAuthServiceServer
	pb.UnimplementedTourServiceServer
	pb.UnimplementedBookingServiceServer
	pb.UnimplementedPaymentServiceServer
	pb.UnimplementedReviewServiceServer
	config      config.Config
	store       db.Querier
	tokenMaker  token.Maker
	oauthConfig *oauth2.Config
}

// NewServer creates a new gRPC server
func NewServer(cfg config.Config, store db.Querier) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(cfg.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	server := &Server{
		config:      cfg,
		store:       store,
		tokenMaker:  tokenMaker,
		oauthConfig: oauthConfig,
	}

	return server, nil
}
