package grpc

import (
	"context"

	appbooking "github.com/irfan-ghzl/pintour/internal/application/booking"
	"github.com/irfan-ghzl/pintour/internal/domain/entity"
	"github.com/irfan-ghzl/pintour/common/interface/middleware"
	pb "github.com/irfan-ghzl/pintour/pb/pintour/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateBooking(ctx context.Context, req *pb.CreateBookingRequest) (*pb.Booking, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}

	var participants []appbooking.ParticipantInput
	for _, p := range req.Participants {
		participants = append(participants, appbooking.ParticipantInput{
			Name:         p.Name,
			IDCardNumber: p.IdCardNumber,
			DateOfBirth:  p.DateOfBirth,
		})
	}

	booking, err := s.bookingService.CreateBooking(ctx, appbooking.CreateBookingInput{
		UserID:          payload.UserID,
		TourPackageID:   req.TourPackageId,
		TourScheduleID:  req.TourScheduleId,
		TravelDate:      req.TravelDate,
		NumParticipants: req.NumParticipants,
		Notes:           req.Notes,
		Participants:    participants,
	})
	if err != nil {
		return nil, err
	}
	return convertBooking(booking), nil
}

func (s *Server) GetBooking(ctx context.Context, req *pb.GetBookingRequest) (*pb.Booking, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}

	booking, err := s.bookingService.GetBooking(ctx, req.Id, payload.UserID, payload.Role)
	if err != nil {
		return nil, err
	}
	return convertBooking(booking), nil
}

func (s *Server) ListBookings(ctx context.Context, req *pb.ListBookingsRequest) (*pb.ListBookingsResponse, error) {
	payload, err := middleware.GetPayload(ctx)
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

	out, err := s.bookingService.ListBookings(ctx, appbooking.ListBookingsInput{
		UserID: payload.UserID,
		Role:   payload.Role,
		Status: entity.BookingStatus(bookingStatusToString(req.Status)),
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}

	var pbBookings []*pb.Booking
	for i := range out.Bookings {
		pbBookings = append(pbBookings, convertBooking(&out.Bookings[i]))
	}
	return &pb.ListBookingsResponse{
		Bookings: pbBookings,
		Pagination: &pb.PaginationResponse{
			Page:       out.Page,
			Limit:      out.Limit,
			Total:      out.Total,
			TotalPages: out.TotalPages,
		},
	}, nil
}

func (s *Server) CancelBooking(ctx context.Context, req *pb.CancelBookingRequest) (*pb.CancelBookingResponse, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}

	booking, err := s.bookingService.CancelBooking(ctx, req.Id, payload.UserID, payload.Role)
	if err != nil {
		return nil, err
	}
	return &pb.CancelBookingResponse{
		Message: "booking cancelled successfully",
		Booking: convertBooking(booking),
	}, nil
}

func (s *Server) UpdateBookingStatus(ctx context.Context, req *pb.UpdateBookingStatusRequest) (*pb.Booking, error) {
	_, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}

	newStatus := entity.BookingStatus(bookingStatusToString(req.Status))
	booking, err := s.bookingService.UpdateStatus(ctx, req.Id, newStatus)
	if err != nil {
		return nil, err
	}
	return convertBooking(booking), nil
}

func bookingStatusToString(s pb.BookingStatus) string {
	switch s {
	case pb.BookingStatus_BOOKING_STATUS_PENDING:
		return "pending"
	case pb.BookingStatus_BOOKING_STATUS_CONFIRMED:
		return "confirmed"
	case pb.BookingStatus_BOOKING_STATUS_CANCELLED:
		return "cancelled"
	case pb.BookingStatus_BOOKING_STATUS_COMPLETED:
		return "completed"
	default:
		return ""
	}
}

func stringToBookingStatus(s string) pb.BookingStatus {
	switch s {
	case "pending":
		return pb.BookingStatus_BOOKING_STATUS_PENDING
	case "confirmed":
		return pb.BookingStatus_BOOKING_STATUS_CONFIRMED
	case "cancelled":
		return pb.BookingStatus_BOOKING_STATUS_CANCELLED
	case "completed":
		return pb.BookingStatus_BOOKING_STATUS_COMPLETED
	default:
		return pb.BookingStatus_BOOKING_STATUS_UNSPECIFIED
	}
}

func convertBooking(b *entity.Booking) *pb.Booking {
	var pbParticipants []*pb.BookingParticipant
	for _, p := range b.Participants {
		dob := ""
		if p.DateOfBirth != nil {
			dob = p.DateOfBirth.Format("2006-01-02")
		}
		pbParticipants = append(pbParticipants, &pb.BookingParticipant{
			Id:           p.ID,
			Name:         p.Name,
			IdCardNumber: p.IDCardNumber,
			DateOfBirth:  dob,
		})
	}

	return &pb.Booking{
		Id:              b.ID,
		BookingCode:     b.BookingCode,
		UserId:          b.UserID,
		TourPackageId:   b.TourPackageID,
		TourTitle:       b.TourTitle,
		TravelDate:      b.TravelDate.Format("2006-01-02"),
		NumParticipants: b.NumParticipants,
		TotalPrice:      b.TotalPrice,
		Status:          stringToBookingStatus(string(b.Status)),
		Notes:           b.Notes,
		Participants:    pbParticipants,
		CreatedAt:       timestamppb.New(b.CreatedAt),
		UpdatedAt:       timestamppb.New(b.UpdatedAt),
	}
}
