package review

import (
	"context"
	"database/sql"
	"time"

	"github.com/irfan-ghzl/pintour/internal/domain/entity"
	"github.com/irfan-ghzl/pintour/internal/domain/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	reviewRepo  repository.ReviewRepository
	bookingRepo repository.BookingRepository
	userRepo    repository.UserRepository
}

func NewService(
	reviewRepo repository.ReviewRepository,
	bookingRepo repository.BookingRepository,
	userRepo repository.UserRepository,
) *Service {
	return &Service{
		reviewRepo:  reviewRepo,
		bookingRepo: bookingRepo,
		userRepo:    userRepo,
	}
}

func (s *Service) CreateReview(ctx context.Context, input CreateReviewInput) (*entity.Review, error) {
	if input.Rating < 1 || input.Rating > 5 {
		return nil, status.Errorf(codes.InvalidArgument, "rating must be between 1 and 5")
	}

	booking, err := s.bookingRepo.GetByID(ctx, input.BookingID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "booking not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot get booking: %v", err)
	}

	if booking.UserID != input.UserID {
		return nil, status.Errorf(codes.PermissionDenied, "access denied")
	}

	if booking.Status != entity.BookingStatusCompleted {
		return nil, status.Errorf(codes.FailedPrecondition, "can only review completed bookings")
	}

	_, err = s.reviewRepo.GetByBooking(ctx, input.UserID, input.BookingID)
	if err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "you have already reviewed this booking")
	}

	review, err := s.reviewRepo.Create(ctx, entity.Review{
		UserID:        input.UserID,
		TourPackageID: input.TourPackageID,
		BookingID:     input.BookingID,
		Rating:        input.Rating,
		Comment:       input.Comment,
		CreatedAt:     time.Now(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create review: %v", err)
	}

	user, _ := s.userRepo.GetByID(ctx, input.UserID)
	if user != nil {
		review.UserName = user.Name
		review.UserAvatar = user.AvatarURL
	}

	return review, nil
}

func (s *Service) ListReviews(ctx context.Context, input ListReviewsInput) (*ListReviewsOutput, error) {
	page := input.Page
	if page < 1 {
		page = 1
	}
	limit := input.Limit
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	reviews, err := s.reviewRepo.ListByTour(ctx, input.TourPackageID, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list reviews: %v", err)
	}

	total, err := s.reviewRepo.CountByTour(ctx, input.TourPackageID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot count reviews: %v", err)
	}

	var totalRating float64
	var reviewsWithUser []ReviewWithUser
	for _, r := range reviews {
		totalRating += float64(r.Rating)
		reviewsWithUser = append(reviewsWithUser, ReviewWithUser{
			ID:            r.ID,
			UserID:        r.UserID,
			UserName:      r.UserName,
			UserAvatar:    r.UserAvatar,
			TourPackageID: r.TourPackageID,
			BookingID:     r.BookingID,
			Rating:        r.Rating,
			Comment:       r.Comment,
			CreatedAt:     r.CreatedAt,
		})
	}

	var avgRating float64
	if len(reviews) > 0 {
		avgRating = totalRating / float64(len(reviews))
	}

	totalPages := int32((total + int64(limit) - 1) / int64(limit))
	return &ListReviewsOutput{
		Reviews:       reviewsWithUser,
		Total:         total,
		TotalPages:    totalPages,
		Page:          page,
		Limit:         limit,
		AverageRating: avgRating,
	}, nil
}

func (s *Service) DeleteReview(ctx context.Context, id int64) error {
	_, err := s.reviewRepo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return status.Errorf(codes.NotFound, "review not found")
		}
		return status.Errorf(codes.Internal, "cannot get review: %v", err)
	}

	if err := s.reviewRepo.Delete(ctx, id); err != nil {
		return status.Errorf(codes.Internal, "cannot delete review: %v", err)
	}
	return nil
}
