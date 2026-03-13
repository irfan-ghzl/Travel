package payment

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/irfan-ghzl/pintour/internal/domain/entity"
	"github.com/irfan-ghzl/pintour/internal/domain/repository"
	infraPayment "github.com/irfan-ghzl/pintour/internal/infrastructure/payment"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	paymentRepo    repository.PaymentRepository
	bookingRepo    repository.BookingRepository
	userRepo       repository.UserRepository
	paymentGateway infraPayment.Gateway
	serverKey      string
}

func NewService(
	paymentRepo repository.PaymentRepository,
	bookingRepo repository.BookingRepository,
	userRepo repository.UserRepository,
	paymentGateway infraPayment.Gateway,
	serverKey string,
) *Service {
	return &Service{
		paymentRepo:    paymentRepo,
		bookingRepo:    bookingRepo,
		userRepo:       userRepo,
		paymentGateway: paymentGateway,
		serverKey:      serverKey,
	}
}

func (s *Service) CreatePayment(ctx context.Context, input CreatePaymentInput) (*CreatePaymentOutput, error) {
	booking, err := s.bookingRepo.GetByID(ctx, input.BookingID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "booking not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot get booking: %v", err)
	}

	if input.UserRole != "admin" && booking.UserID != input.UserID {
		return nil, status.Errorf(codes.PermissionDenied, "access denied")
	}

	user, _ := s.userRepo.GetByID(ctx, booking.UserID)

	var userName, userEmail, userPhone string
	if user != nil {
		userName = user.Name
		userEmail = user.Email
		userPhone = user.Phone
	}

	result, err := s.paymentGateway.CreateTransaction(ctx, entity.PaymentRequest{
		BookingCode: booking.BookingCode,
		Amount:      booking.TotalPrice,
		UserName:    userName,
		UserEmail:   userEmail,
		UserPhone:   userPhone,
	})

	paymentToken := ""
	paymentURL := ""
	if err == nil && result != nil {
		paymentToken = result.Token
		paymentURL = result.URL
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	p, err := s.paymentRepo.Create(ctx, repository.CreatePaymentParams{
		BookingID:     booking.ID,
		PaymentMethod: input.PaymentMethod,
		Amount:        booking.TotalPrice,
		Status:        entity.PaymentStatusPending,
		PaymentToken:  paymentToken,
		PaymentURL:    paymentURL,
		ExpiresAt:     &expiresAt,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create payment: %v", err)
	}

	p.BookingCode = booking.BookingCode
	return &CreatePaymentOutput{
		Payment:      p,
		PaymentURL:   paymentURL,
		PaymentToken: paymentToken,
	}, nil
}

func (s *Service) GetPayment(ctx context.Context, bookingID, userID int64, role string) (*entity.Payment, error) {
	p, err := s.paymentRepo.GetByBookingID(ctx, bookingID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "payment not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot get payment: %v", err)
	}

	booking, _ := s.bookingRepo.GetByID(ctx, p.BookingID)
	if booking != nil {
		if role != "admin" && booking.UserID != userID {
			return nil, status.Errorf(codes.PermissionDenied, "access denied")
		}
		p.BookingCode = booking.BookingCode
	}

	return p, nil
}

func (s *Service) HandleNotification(ctx context.Context, orderID, transactionStatus, fraudStatus, grossAmount, signatureKey string) error {
	signatureInput := orderID + transactionStatus + fraudStatus + grossAmount + s.serverKey
	h := sha512.New()
	h.Write([]byte(signatureInput))
	expectedSig := fmt.Sprintf("%x", h.Sum(nil))

	if !strings.EqualFold(expectedSig, signatureKey) {
		return status.Errorf(codes.Unauthenticated, "invalid signature")
	}

	booking, err := s.bookingRepo.GetByCode(ctx, orderID)
	if err != nil {
		return status.Errorf(codes.NotFound, "booking not found")
	}

	p, err := s.paymentRepo.GetByBookingID(ctx, booking.ID)
	if err != nil {
		return status.Errorf(codes.NotFound, "payment not found")
	}

	newStatus := entity.PaymentStatusPending
	switch transactionStatus {
	case "capture", "settlement":
		if fraudStatus == "accept" || fraudStatus == "" {
			newStatus = entity.PaymentStatusPaid
		} else {
			newStatus = entity.PaymentStatusFailed
		}
	case "cancel", "deny":
		newStatus = entity.PaymentStatusFailed
	case "expire":
		newStatus = entity.PaymentStatusExpired
	}

	_, err = s.paymentRepo.UpdateStatus(ctx, p.ID, newStatus)
	if err != nil {
		return status.Errorf(codes.Internal, "cannot update payment: %v", err)
	}

	if newStatus == entity.PaymentStatusPaid {
		_, _ = s.bookingRepo.UpdateStatus(ctx, booking.ID, entity.BookingStatusConfirmed)
	} else if newStatus == entity.PaymentStatusFailed || newStatus == entity.PaymentStatusExpired {
		_, _ = s.bookingRepo.UpdateStatus(ctx, booking.ID, entity.BookingStatusCancelled)
	}

	return nil
}

func (s *Service) ListPayments(ctx context.Context, input ListPaymentsInput) (*ListPaymentsOutput, error) {
	page := input.Page
	if page < 1 {
		page = 1
	}
	limit := input.Limit
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit
	statusStr := string(input.Status)

	payments, err := s.paymentRepo.List(ctx, statusStr, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list payments: %v", err)
	}

	total, err := s.paymentRepo.Count(ctx, statusStr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot count payments: %v", err)
	}

	for i := range payments {
		booking, _ := s.bookingRepo.GetByID(ctx, payments[i].BookingID)
		if booking != nil {
			payments[i].BookingCode = booking.BookingCode
		}
	}

	totalPages := int32((total + int64(limit) - 1) / int64(limit))
	return &ListPaymentsOutput{
		Payments:   payments,
		Total:      total,
		TotalPages: totalPages,
		Page:       page,
		Limit:      limit,
	}, nil
}
