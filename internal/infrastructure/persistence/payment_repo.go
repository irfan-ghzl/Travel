package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	db "github.com/irfan-ghzl/pintour/db/sqlc"
	"github.com/irfan-ghzl/pintour/internal/domain/entity"
	"github.com/irfan-ghzl/pintour/internal/domain/repository"
)

type paymentRepository struct {
	q db.Querier
}

func NewPaymentRepository(q db.Querier) repository.PaymentRepository {
	return &paymentRepository{q: q}
}

func (r *paymentRepository) Create(ctx context.Context, params repository.CreatePaymentParams) (*entity.Payment, error) {
	var expiresAt sql.NullTime
	if t, ok := params.ExpiresAt.(*time.Time); ok && t != nil {
		expiresAt = sql.NullTime{Time: *t, Valid: true}
	}

	result, err := r.q.CreatePayment(ctx, db.CreatePaymentParams{
		BookingID:     params.BookingID,
		PaymentMethod: string(params.PaymentMethod),
		Amount:        fmt.Sprintf("%.2f", params.Amount),
		Status:        string(params.Status),
		PaymentToken:  params.PaymentToken,
		PaymentUrl:    params.PaymentURL,
		ExpiresAt:     expiresAt,
	})
	if err != nil {
		return nil, err
	}
	return toEntityPayment(result), nil
}

func (r *paymentRepository) GetByBookingID(ctx context.Context, bookingID int64) (*entity.Payment, error) {
	result, err := r.q.GetPaymentByBookingID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	return toEntityPayment(result), nil
}

func (r *paymentRepository) GetByID(ctx context.Context, id int64) (*entity.Payment, error) {
	result, err := r.q.GetPaymentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toEntityPayment(result), nil
}

func (r *paymentRepository) UpdateStatus(ctx context.Context, id int64, status entity.PaymentStatus) (*entity.Payment, error) {
	result, err := r.q.UpdatePaymentStatus(ctx, db.UpdatePaymentStatusParams{
		ID:     id,
		Status: string(status),
	})
	if err != nil {
		return nil, err
	}
	return toEntityPayment(result), nil
}

func (r *paymentRepository) UpdateToken(ctx context.Context, id int64, token, url string) (*entity.Payment, error) {
	result, err := r.q.UpdatePaymentToken(ctx, db.UpdatePaymentTokenParams{
		ID:           id,
		PaymentToken: token,
		PaymentUrl:   url,
	})
	if err != nil {
		return nil, err
	}
	return toEntityPayment(result), nil
}

func (r *paymentRepository) List(ctx context.Context, status string, limit, offset int32) ([]entity.Payment, error) {
	results, err := r.q.ListPayments(ctx, db.ListPaymentsParams{
		Column1: status,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, err
	}
	var payments []entity.Payment
	for _, p := range results {
		payments = append(payments, *toEntityPayment(p))
	}
	return payments, nil
}

func (r *paymentRepository) Count(ctx context.Context, status string) (int64, error) {
	return r.q.CountPayments(ctx, status)
}

func toEntityPayment(p db.Payment) *entity.Payment {
	amount, _ := strconv.ParseFloat(p.Amount, 64)
	payment := &entity.Payment{
		ID:            p.ID,
		BookingID:     p.BookingID,
		PaymentMethod: entity.PaymentMethod(p.PaymentMethod),
		Amount:        amount,
		Status:        entity.PaymentStatus(p.Status),
		PaymentToken:  p.PaymentToken,
		PaymentURL:    p.PaymentUrl,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
	if p.PaidAt.Valid {
		payment.PaidAt = &p.PaidAt.Time
	}
	if p.ExpiresAt.Valid {
		payment.ExpiresAt = &p.ExpiresAt.Time
	}
	return payment
}
