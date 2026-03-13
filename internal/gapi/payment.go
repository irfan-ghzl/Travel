package gapi

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	db "github.com/irfan-ghzl/pintour/internal/db/sqlc"
	"github.com/irfan-ghzl/pintour/internal/middleware"
	pb "github.com/irfan-ghzl/pintour/pb/pintour/v1"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// paymentMethodToString converts proto PaymentMethod to DB string
func paymentMethodToString(m pb.PaymentMethod) string {
	switch m {
	case pb.PaymentMethod_PAYMENT_METHOD_BANK_TRANSFER:
		return "bank_transfer"
	case pb.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD:
		return "credit_card"
	case pb.PaymentMethod_PAYMENT_METHOD_GOPAY:
		return "gopay"
	case pb.PaymentMethod_PAYMENT_METHOD_OVO:
		return "ovo"
	case pb.PaymentMethod_PAYMENT_METHOD_DANA:
		return "dana"
	case pb.PaymentMethod_PAYMENT_METHOD_QRIS:
		return "qris"
	default:
		return "bank_transfer"
	}
}

// stringToPaymentMethod converts DB string to proto PaymentMethod
func stringToPaymentMethod(s string) pb.PaymentMethod {
	switch s {
	case "bank_transfer":
		return pb.PaymentMethod_PAYMENT_METHOD_BANK_TRANSFER
	case "credit_card":
		return pb.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD
	case "gopay":
		return pb.PaymentMethod_PAYMENT_METHOD_GOPAY
	case "ovo":
		return pb.PaymentMethod_PAYMENT_METHOD_OVO
	case "dana":
		return pb.PaymentMethod_PAYMENT_METHOD_DANA
	case "qris":
		return pb.PaymentMethod_PAYMENT_METHOD_QRIS
	default:
		return pb.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED
	}
}

// stringToPaymentStatus converts DB string to proto PaymentStatus
func stringToPaymentStatus(s string) pb.PaymentStatus {
	switch s {
	case "pending":
		return pb.PaymentStatus_PAYMENT_STATUS_PENDING
	case "paid":
		return pb.PaymentStatus_PAYMENT_STATUS_PAID
	case "failed":
		return pb.PaymentStatus_PAYMENT_STATUS_FAILED
	case "refunded":
		return pb.PaymentStatus_PAYMENT_STATUS_REFUNDED
	case "expired":
		return pb.PaymentStatus_PAYMENT_STATUS_EXPIRED
	default:
		return pb.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
	}
}

// paymentStatusToString converts proto PaymentStatus to DB string
func paymentStatusToString(s pb.PaymentStatus) string {
	switch s {
	case pb.PaymentStatus_PAYMENT_STATUS_PENDING:
		return "pending"
	case pb.PaymentStatus_PAYMENT_STATUS_PAID:
		return "paid"
	case pb.PaymentStatus_PAYMENT_STATUS_FAILED:
		return "failed"
	case pb.PaymentStatus_PAYMENT_STATUS_REFUNDED:
		return "refunded"
	case pb.PaymentStatus_PAYMENT_STATUS_EXPIRED:
		return "expired"
	default:
		return ""
	}
}

// convertPayment converts db.Payment to pb.Payment
func convertPayment(p db.Payment, bookingCode string) *pb.Payment {
	amount, _ := strconv.ParseFloat(p.Amount, 64)

	pbPayment := &pb.Payment{
		Id:            p.ID,
		BookingId:     p.BookingID,
		BookingCode:   bookingCode,
		PaymentMethod: stringToPaymentMethod(p.PaymentMethod),
		Amount:        amount,
		Status:        stringToPaymentStatus(p.Status),
		PaymentToken:  p.PaymentToken,
		PaymentUrl:    p.PaymentUrl,
		CreatedAt:     timestamppb.New(p.CreatedAt),
		UpdatedAt:     timestamppb.New(p.UpdatedAt),
	}
	if p.PaidAt.Valid {
		pbPayment.PaidAt = timestamppb.New(p.PaidAt.Time)
	}
	if p.ExpiresAt.Valid {
		pbPayment.ExpiresAt = timestamppb.New(p.ExpiresAt.Time)
	}
	return pbPayment
}

// CreatePayment creates a payment for a booking via Midtrans
func (s *Server) CreatePayment(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.CreatePaymentResponse, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}

	booking, err := s.store.GetBooking(ctx, req.BookingId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "booking not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot get booking: %v", err)
	}

	if payload.Role != "admin" && booking.UserID != payload.UserID {
		return nil, status.Errorf(codes.PermissionDenied, "access denied")
	}

	// Create Midtrans Snap client
	env := midtrans.Sandbox
	if s.config.MidtransIsProduction {
		env = midtrans.Production
	}
	snapClient := snap.Client{}
	snapClient.New(s.config.MidtransServerKey, env)

	amount, _ := strconv.ParseFloat(booking.TotalPrice, 64)
	user, _ := s.store.GetUserByID(ctx, booking.UserID)

	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  booking.BookingCode,
			GrossAmt: int64(amount),
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: user.Name,
			Email: user.Email,
			Phone: user.Phone,
		},
	}

	snapResp, snapErr := snapClient.CreateTransaction(snapReq)

	expiresAt := sql.NullTime{Time: time.Now().Add(24 * time.Hour), Valid: true}
	paymentToken := ""
	paymentURL := ""
	if snapErr == nil && snapResp != nil {
		paymentToken = snapResp.Token
		paymentURL = snapResp.RedirectURL
	}

	payment, err := s.store.CreatePayment(ctx, db.CreatePaymentParams{
		BookingID:     booking.ID,
		PaymentMethod: paymentMethodToString(req.PaymentMethod),
		Amount:        booking.TotalPrice,
		Status:        "pending",
		PaymentToken:  paymentToken,
		PaymentUrl:    paymentURL,
		ExpiresAt:     expiresAt,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create payment: %v", err)
	}

	pbPayment := convertPayment(payment, booking.BookingCode)
	return &pb.CreatePaymentResponse{
		Payment:      pbPayment,
		PaymentUrl:   paymentURL,
		PaymentToken: paymentToken,
	}, nil
}

// GetPayment returns the payment for a booking
func (s *Server) GetPayment(ctx context.Context, req *pb.GetPaymentRequest) (*pb.Payment, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}

	payment, err := s.store.GetPaymentByBookingID(ctx, req.BookingId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "payment not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot get payment: %v", err)
	}

	booking, _ := s.store.GetBooking(ctx, payment.BookingID)
	if payload.Role != "admin" && booking.UserID != payload.UserID {
		return nil, status.Errorf(codes.PermissionDenied, "access denied")
	}

	return convertPayment(payment, booking.BookingCode), nil
}

// PaymentNotification handles Midtrans payment webhook notifications
func (s *Server) PaymentNotification(ctx context.Context, req *pb.PaymentNotificationRequest) (*pb.PaymentNotificationResponse, error) {
	// Verify signature
	signatureInput := req.OrderId + req.TransactionStatus + req.FraudStatus + req.GrossAmount + s.config.MidtransServerKey
	h := sha512.New()
	h.Write([]byte(signatureInput))
	expectedSig := fmt.Sprintf("%x", h.Sum(nil))

	if !strings.EqualFold(expectedSig, req.SignatureKey) {
		return nil, status.Errorf(codes.Unauthenticated, "invalid signature")
	}

	booking, err := s.store.GetBookingByCode(ctx, req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "booking not found")
	}

	payment, err := s.store.GetPaymentByBookingID(ctx, booking.ID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "payment not found")
	}

	newStatus := "pending"
	switch req.TransactionStatus {
	case "capture", "settlement":
		if req.FraudStatus == "accept" || req.FraudStatus == "" {
			newStatus = "paid"
		} else {
			newStatus = "failed"
		}
	case "cancel", "deny", "expire":
		newStatus = "failed"
		if req.TransactionStatus == "expire" {
			newStatus = "expired"
		}
	}

	_, err = s.store.UpdatePaymentStatus(ctx, db.UpdatePaymentStatusParams{
		ID:     payment.ID,
		Status: newStatus,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot update payment: %v", err)
	}

	// Update booking status based on payment
	if newStatus == "paid" {
		_, _ = s.store.UpdateBookingStatus(ctx, db.UpdateBookingStatusParams{
			ID:     booking.ID,
			Status: "confirmed",
		})
	} else if newStatus == "failed" || newStatus == "expired" {
		_, _ = s.store.UpdateBookingStatus(ctx, db.UpdateBookingStatusParams{
			ID:     booking.ID,
			Status: "cancelled",
		})
	}

	return &pb.PaymentNotificationResponse{Message: "notification processed"}, nil
}

// ListPayments lists all payments (admin only)
func (s *Server) ListPayments(ctx context.Context, req *pb.ListPaymentsRequest) (*pb.ListPaymentsResponse, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}
	if payload.Role != "admin" {
		return nil, status.Errorf(codes.PermissionDenied, "only admins can list all payments")
	}

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
	offset := (page - 1) * limit
	statusStr := paymentStatusToString(req.Status)

	payments, err := s.store.ListPayments(ctx, db.ListPaymentsParams{
		Column1: statusStr,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list payments: %v", err)
	}

	total, err := s.store.CountPayments(ctx, statusStr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot count payments: %v", err)
	}

	var pbPayments []*pb.Payment
	for _, p := range payments {
		booking, _ := s.store.GetBooking(ctx, p.BookingID)
		pbPayments = append(pbPayments, convertPayment(p, booking.BookingCode))
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	return &pb.ListPaymentsResponse{
		Payments: pbPayments,
		Pagination: &pb.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: int32(totalPages),
		},
	}, nil
}
