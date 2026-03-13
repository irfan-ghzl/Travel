package grpc

import (
	"github.com/irfan-ghzl/pintour/internal/application/auth"
	appbooking "github.com/irfan-ghzl/pintour/internal/application/booking"
	apppayment "github.com/irfan-ghzl/pintour/internal/application/payment"
	appreview "github.com/irfan-ghzl/pintour/internal/application/review"
	apptour "github.com/irfan-ghzl/pintour/internal/application/tour"
	pb "github.com/irfan-ghzl/pintour/pb/pintour/v1"
)

// Server is the gRPC server that delegates to application services.
type Server struct {
	pb.UnimplementedAuthServiceServer
	pb.UnimplementedTourServiceServer
	pb.UnimplementedBookingServiceServer
	pb.UnimplementedPaymentServiceServer
	pb.UnimplementedReviewServiceServer

	authService    *auth.Service
	tourService    *apptour.Service
	bookingService *appbooking.Service
	paymentService *apppayment.Service
	reviewService  *appreview.Service
}

// NewServer creates a new gRPC server.
func NewServer(
	authService *auth.Service,
	tourService *apptour.Service,
	bookingService *appbooking.Service,
	paymentService *apppayment.Service,
	reviewService *appreview.Service,
) *Server {
	return &Server{
		authService:    authService,
		tourService:    tourService,
		bookingService: bookingService,
		paymentService: paymentService,
		reviewService:  reviewService,
	}
}
