package gapi

import (
	"context"
	"database/sql"
	"strconv"

	db "github.com/irfan-ghzl/pintour/internal/db/sqlc"
	"github.com/irfan-ghzl/pintour/internal/middleware"
	pb "github.com/irfan-ghzl/pintour/pb/pintour/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateReview creates a review for a completed tour
func (s *Server) CreateReview(ctx context.Context, req *pb.CreateReviewRequest) (*pb.Review, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}

	if req.Rating < 1 || req.Rating > 5 {
		return nil, status.Errorf(codes.InvalidArgument, "rating must be between 1 and 5")
	}

	// Check that the user has a completed booking for this tour
	booking, err := s.store.GetBooking(ctx, req.BookingId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "booking not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot get booking: %v", err)
	}

	if booking.UserID != payload.UserID {
		return nil, status.Errorf(codes.PermissionDenied, "access denied")
	}

	if booking.Status != "completed" {
		return nil, status.Errorf(codes.FailedPrecondition, "can only review completed bookings")
	}

	// Check if already reviewed
	_, err = s.store.GetReviewByBooking(ctx, db.GetReviewByBookingParams{
		UserID:    payload.UserID,
		BookingID: req.BookingId,
	})
	if err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "you have already reviewed this booking")
	}

	review, err := s.store.CreateReview(ctx, db.CreateReviewParams{
		UserID:        payload.UserID,
		TourPackageID: req.TourPackageId,
		BookingID:     req.BookingId,
		Rating:        req.Rating,
		Comment:       req.Comment,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create review: %v", err)
	}

	user, _ := s.store.GetUserByID(ctx, payload.UserID)

	return &pb.Review{
		Id:            review.ID,
		UserId:        review.UserID,
		UserName:      user.Name,
		UserAvatar:    user.AvatarUrl,
		TourPackageId: review.TourPackageID,
		BookingId:     review.BookingID,
		Rating:        review.Rating,
		Comment:       review.Comment,
		CreatedAt:     timestamppb.New(review.CreatedAt),
	}, nil
}

// ListReviews returns reviews for a tour package
func (s *Server) ListReviews(ctx context.Context, req *pb.ListReviewsRequest) (*pb.ListReviewsResponse, error) {
	page := int32(1)
	limit := int32(10)
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = req.Pagination.Page
		}
		if req.Pagination.Limit > 0 {
			limit = req.Pagination.Limit
		}
	}
	offset := (page - 1) * limit

	reviews, err := s.store.ListReviewsByTour(ctx, db.ListReviewsByTourParams{
		TourPackageID: req.TourPackageId,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list reviews: %v", err)
	}

	total, err := s.store.CountReviewsByTour(ctx, req.TourPackageId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot count reviews: %v", err)
	}

	avgRatingStr, _ := s.store.GetAverageRating(ctx, req.TourPackageId)
	avgRating, _ := strconv.ParseFloat(avgRatingStr, 64)

	var pbReviews []*pb.Review
	for _, r := range reviews {
		pbReviews = append(pbReviews, &pb.Review{
			Id:            r.ID,
			UserId:        r.UserID,
			UserName:      r.UserName,
			UserAvatar:    r.UserAvatar,
			TourPackageId: r.TourPackageID,
			BookingId:     r.BookingID,
			Rating:        r.Rating,
			Comment:       r.Comment,
			CreatedAt:     timestamppb.New(r.CreatedAt),
		})
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	return &pb.ListReviewsResponse{
		Reviews:       pbReviews,
		AverageRating: avgRating,
		Pagination: &pb.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: int32(totalPages),
		},
	}, nil
}

// DeleteReview deletes a review (admin only)
func (s *Server) DeleteReview(ctx context.Context, req *pb.DeleteReviewRequest) (*pb.DeleteReviewResponse, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}
	if payload.Role != "admin" {
		return nil, status.Errorf(codes.PermissionDenied, "only admins can delete reviews")
	}

	_, err = s.store.GetReview(ctx, req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "review not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot get review: %v", err)
	}

	err = s.store.DeleteReview(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot delete review: %v", err)
	}

	return &pb.DeleteReviewResponse{Message: "review deleted successfully"}, nil
}
