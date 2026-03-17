package persistence

import (
	"context"

	db "github.com/irfan-ghzl/pintour/db/sqlc"
	"github.com/irfan-ghzl/pintour/internal/domain/entity"
	"github.com/irfan-ghzl/pintour/internal/domain/repository"
)

type reviewRepository struct {
	q db.Querier
}

func NewReviewRepository(q db.Querier) repository.ReviewRepository {
	return &reviewRepository{q: q}
}

func (r *reviewRepository) Create(ctx context.Context, rev entity.Review) (*entity.Review, error) {
	result, err := r.q.CreateReview(ctx, db.CreateReviewParams{
		UserID:        rev.UserID,
		TourPackageID: rev.TourPackageID,
		BookingID:     rev.BookingID,
		Rating:        rev.Rating,
		Comment:       rev.Comment,
	})
	if err != nil {
		return nil, err
	}
	return toEntityReview(result, rev.UserName, rev.UserAvatar), nil
}

func (r *reviewRepository) GetByID(ctx context.Context, id int64) (*entity.Review, error) {
	result, err := r.q.GetReview(ctx, id)
	if err != nil {
		return nil, err
	}
	return toEntityReview(result, "", ""), nil
}

func (r *reviewRepository) GetByBooking(ctx context.Context, userID, bookingID int64) (*entity.Review, error) {
	result, err := r.q.GetReviewByBooking(ctx, db.GetReviewByBookingParams{
		UserID:    userID,
		BookingID: bookingID,
	})
	if err != nil {
		return nil, err
	}
	return toEntityReview(result, "", ""), nil
}

func (r *reviewRepository) ListByTour(ctx context.Context, tourID int64, limit, offset int32) ([]entity.Review, error) {
	results, err := r.q.ListReviewsByTour(ctx, db.ListReviewsByTourParams{
		TourPackageID: tourID,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, err
	}
	var reviews []entity.Review
	for _, row := range results {
		reviews = append(reviews, entity.Review{
			ID:            row.ID,
			UserID:        row.UserID,
			UserName:      row.UserName,
			UserAvatar:    row.UserAvatar,
			TourPackageID: row.TourPackageID,
			BookingID:     row.BookingID,
			Rating:        row.Rating,
			Comment:       row.Comment,
			CreatedAt:     row.CreatedAt,
		})
	}
	return reviews, nil
}

func (r *reviewRepository) CountByTour(ctx context.Context, tourID int64) (int64, error) {
	return r.q.CountReviewsByTour(ctx, tourID)
}

func (r *reviewRepository) Delete(ctx context.Context, id int64) error {
	return r.q.DeleteReview(ctx, id)
}

func toEntityReview(r db.Review, userName, userAvatar string) *entity.Review {
	return &entity.Review{
		ID:            r.ID,
		UserID:        r.UserID,
		UserName:      userName,
		UserAvatar:    userAvatar,
		TourPackageID: r.TourPackageID,
		BookingID:     r.BookingID,
		Rating:        r.Rating,
		Comment:       r.Comment,
		CreatedAt:     r.CreatedAt,
	}
}
