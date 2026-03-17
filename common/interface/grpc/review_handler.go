package grpc

import (
	"context"

	appreview "github.com/irfan-ghzl/pintour/internal/application/review"
	"github.com/irfan-ghzl/pintour/common/interface/middleware"
	pb "github.com/irfan-ghzl/pintour/pb/pintour/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateReview(ctx context.Context, req *pb.CreateReviewRequest) (*pb.Review, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}

	review, err := s.reviewService.CreateReview(ctx, appreview.CreateReviewInput{
		UserID:        payload.UserID,
		TourPackageID: req.TourPackageId,
		BookingID:     req.BookingId,
		Rating:        req.Rating,
		Comment:       req.Comment,
	})
	if err != nil {
		return nil, err
	}

	return &pb.Review{
		Id:            review.ID,
		UserId:        review.UserID,
		UserName:      review.UserName,
		UserAvatar:    review.UserAvatar,
		TourPackageId: review.TourPackageID,
		BookingId:     review.BookingID,
		Rating:        review.Rating,
		Comment:       review.Comment,
		CreatedAt:     timestamppb.New(review.CreatedAt),
	}, nil
}

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

	out, err := s.reviewService.ListReviews(ctx, appreview.ListReviewsInput{
		TourPackageID: req.TourPackageId,
		Page:          page,
		Limit:         limit,
	})
	if err != nil {
		return nil, err
	}

	var pbReviews []*pb.Review
	for _, r := range out.Reviews {
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

	return &pb.ListReviewsResponse{
		Reviews:       pbReviews,
		AverageRating: out.AverageRating,
		Pagination: &pb.PaginationResponse{
			Page:       out.Page,
			Limit:      out.Limit,
			Total:      out.Total,
			TotalPages: out.TotalPages,
		},
	}, nil
}

func (s *Server) DeleteReview(ctx context.Context, req *pb.DeleteReviewRequest) (*pb.DeleteReviewResponse, error) {
	_, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.reviewService.DeleteReview(ctx, req.Id); err != nil {
		return nil, err
	}
	return &pb.DeleteReviewResponse{Message: "review deleted successfully"}, nil
}
