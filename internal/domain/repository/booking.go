package repository

import (
	"context"

	"github.com/irfan-ghzl/pintour/internal/domain/entity"
)

type CreateBookingParams struct {
	BookingCode     string
	UserID          int64
	TourPackageID   int64
	TourScheduleID  *int64
	TravelDate      interface{}
	NumParticipants int32
	TotalPrice      float64
	Notes           string
}

type BookingRepository interface {
	Create(ctx context.Context, params CreateBookingParams) (*entity.Booking, error)
	GetByID(ctx context.Context, id int64) (*entity.Booking, error)
	GetByCode(ctx context.Context, code string) (*entity.Booking, error)
	ListByUser(ctx context.Context, userID int64, status string, limit, offset int32) ([]entity.Booking, error)
	CountByUser(ctx context.Context, userID int64, status string) (int64, error)
	ListAll(ctx context.Context, status string, limit, offset int32) ([]entity.Booking, error)
	CountAll(ctx context.Context, status string) (int64, error)
	UpdateStatus(ctx context.Context, id int64, status entity.BookingStatus) (*entity.Booking, error)

	CreateParticipant(ctx context.Context, p entity.BookingParticipant) (*entity.BookingParticipant, error)
	ListParticipants(ctx context.Context, bookingID int64) ([]entity.BookingParticipant, error)
}
