package repository

import (
	"context"

	"github.com/irfan-ghzl/pintour/internal/domain/entity"
)

type ReviewRepository interface {
	Create(ctx context.Context, r entity.Review) (*entity.Review, error)
	GetByID(ctx context.Context, id int64) (*entity.Review, error)
	GetByBooking(ctx context.Context, userID, bookingID int64) (*entity.Review, error)
	ListByTour(ctx context.Context, tourID int64, limit, offset int32) ([]entity.Review, error)
	CountByTour(ctx context.Context, tourID int64) (int64, error)
	Delete(ctx context.Context, id int64) error
}
