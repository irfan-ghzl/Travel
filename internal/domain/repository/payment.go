package repository

import (
	"context"

	"github.com/irfan-ghzl/pintour/internal/domain/entity"
)

type CreatePaymentParams struct {
	BookingID     int64
	PaymentMethod entity.PaymentMethod
	Amount        float64
	Status        entity.PaymentStatus
	PaymentToken  string
	PaymentURL    string
	ExpiresAt     interface{}
}

type PaymentRepository interface {
	Create(ctx context.Context, params CreatePaymentParams) (*entity.Payment, error)
	GetByBookingID(ctx context.Context, bookingID int64) (*entity.Payment, error)
	GetByID(ctx context.Context, id int64) (*entity.Payment, error)
	UpdateStatus(ctx context.Context, id int64, status entity.PaymentStatus) (*entity.Payment, error)
	UpdateToken(ctx context.Context, id int64, token, url string) (*entity.Payment, error)
	List(ctx context.Context, status string, limit, offset int32) ([]entity.Payment, error)
	Count(ctx context.Context, status string) (int64, error)
}
