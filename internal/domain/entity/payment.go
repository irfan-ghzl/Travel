package entity

import "time"

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusFailed   PaymentStatus = "failed"
	PaymentStatusRefunded PaymentStatus = "refunded"
	PaymentStatusExpired  PaymentStatus = "expired"
)

type PaymentMethod string

const (
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodCreditCard   PaymentMethod = "credit_card"
	PaymentMethodGopay        PaymentMethod = "gopay"
	PaymentMethodOvo          PaymentMethod = "ovo"
	PaymentMethodDana         PaymentMethod = "dana"
	PaymentMethodQris         PaymentMethod = "qris"
)

type Payment struct {
	ID            int64
	BookingID     int64
	BookingCode   string
	PaymentMethod PaymentMethod
	Amount        float64
	Status        PaymentStatus
	PaymentToken  string
	PaymentURL    string
	PaidAt        *time.Time
	ExpiresAt     *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type PaymentRequest struct {
	BookingCode string
	Amount      float64
	UserName    string
	UserEmail   string
	UserPhone   string
}

type PaymentResult struct {
	Token string
	URL   string
}
