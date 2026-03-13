package payment

import (
	"github.com/irfan-ghzl/pintour/internal/domain/entity"
)

type CreatePaymentInput struct {
	UserID        int64
	UserRole      string
	BookingID     int64
	PaymentMethod entity.PaymentMethod
}

type CreatePaymentOutput struct {
	Payment      *entity.Payment
	PaymentURL   string
	PaymentToken string
}

type ListPaymentsInput struct {
	Status entity.PaymentStatus
	Page   int32
	Limit  int32
}

type ListPaymentsOutput struct {
	Payments   []entity.Payment
	Total      int64
	TotalPages int32
	Page       int32
	Limit      int32
}
