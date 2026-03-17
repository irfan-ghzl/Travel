package middleware

import (
	"context"
	"strings"

	"github.com/irfan-ghzl/pintour/common/token"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	authorizationHeader = "authorization"
	authorizationBearer = "bearer"
)

type contextKey string

const PayloadKey contextKey = "payload"

var publicMethods = map[string]bool{
	"/pintour.v1.AuthService/Register":               true,
	"/pintour.v1.AuthService/Login":                  true,
	"/pintour.v1.AuthService/GoogleLogin":            true,
	"/pintour.v1.TourService/ListTourPackages":       true,
	"/pintour.v1.TourService/GetTourPackage":         true,
	"/pintour.v1.TourService/ListDestinations":       true,
	"/pintour.v1.TourService/ListTourSchedules":      true,
	"/pintour.v1.ReviewService/ListReviews":          true,
	"/pintour.v1.PaymentService/PaymentNotification": true,
}

func AuthInterceptor(tokenMaker token.Maker) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		values := md.Get(authorizationHeader)
		if len(values) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		authHeader := values[0]
		fields := strings.Fields(authHeader)
		if len(fields) < 2 {
			return nil, status.Errorf(codes.Unauthenticated, "invalid authorization header format")
		}

		authType := strings.ToLower(fields[0])
		if authType != authorizationBearer {
			return nil, status.Errorf(codes.Unauthenticated, "unsupported authorization type %s", authType)
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "access token is invalid: %s", err)
		}

		ctx = context.WithValue(ctx, PayloadKey, payload)
		return handler(ctx, req)
	}
}

func GetPayload(ctx context.Context) (*token.Payload, error) {
	payload, ok := ctx.Value(PayloadKey).(*token.Payload)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}
	return payload, nil
}
