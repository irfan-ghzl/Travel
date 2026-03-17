package grpc

import (
	"context"

	apppayment "github.com/irfan-ghzl/pintour/internal/application/payment"
	"github.com/irfan-ghzl/pintour/internal/domain/entity"
	"github.com/irfan-ghzl/pintour/common/interface/middleware"
	pb "github.com/irfan-ghzl/pintour/pb/pintour/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreatePayment(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.CreatePaymentResponse, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}

	out, err := s.paymentService.CreatePayment(ctx, apppayment.CreatePaymentInput{
		UserID:        payload.UserID,
		UserRole:      payload.Role,
		BookingID:     req.BookingId,
		PaymentMethod: entity.PaymentMethod(paymentMethodToString(req.PaymentMethod)),
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreatePaymentResponse{
		Payment:      convertPayment(out.Payment),
		PaymentUrl:   out.PaymentURL,
		PaymentToken: out.PaymentToken,
	}, nil
}

func (s *Server) GetPayment(ctx context.Context, req *pb.GetPaymentRequest) (*pb.Payment, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}

	p, err := s.paymentService.GetPayment(ctx, req.BookingId, payload.UserID, payload.Role)
	if err != nil {
		return nil, err
	}
	return convertPayment(p), nil
}

func (s *Server) PaymentNotification(ctx context.Context, req *pb.PaymentNotificationRequest) (*pb.PaymentNotificationResponse, error) {
	if err := s.paymentService.HandleNotification(
		ctx,
		req.OrderId,
		req.TransactionStatus,
		req.FraudStatus,
		req.GrossAmount,
		req.SignatureKey,
	); err != nil {
		return nil, err
	}
	return &pb.PaymentNotificationResponse{Message: "notification processed"}, nil
}

func (s *Server) ListPayments(ctx context.Context, req *pb.ListPaymentsRequest) (*pb.ListPaymentsResponse, error) {
	_, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
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

	out, err := s.paymentService.ListPayments(ctx, apppayment.ListPaymentsInput{
		Status: entity.PaymentStatus(paymentStatusToString(req.Status)),
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}

	var pbPayments []*pb.Payment
	for i := range out.Payments {
		pbPayments = append(pbPayments, convertPayment(&out.Payments[i]))
	}
	return &pb.ListPaymentsResponse{
		Payments: pbPayments,
		Pagination: &pb.PaginationResponse{
			Page:       out.Page,
			Limit:      out.Limit,
			Total:      out.Total,
			TotalPages: out.TotalPages,
		},
	}, nil
}

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

func convertPayment(p *entity.Payment) *pb.Payment {
	pbPayment := &pb.Payment{
		Id:            p.ID,
		BookingId:     p.BookingID,
		BookingCode:   p.BookingCode,
		PaymentMethod: stringToPaymentMethod(string(p.PaymentMethod)),
		Amount:        p.Amount,
		Status:        stringToPaymentStatus(string(p.Status)),
		PaymentToken:  p.PaymentToken,
		PaymentUrl:    p.PaymentURL,
		CreatedAt:     timestamppb.New(p.CreatedAt),
		UpdatedAt:     timestamppb.New(p.UpdatedAt),
	}
	if p.PaidAt != nil {
		pbPayment.PaidAt = timestamppb.New(*p.PaidAt)
	}
	if p.ExpiresAt != nil {
		pbPayment.ExpiresAt = timestamppb.New(*p.ExpiresAt)
	}
	return pbPayment
}
