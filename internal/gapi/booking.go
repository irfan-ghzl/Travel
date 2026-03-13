package gapi

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	db "github.com/irfan-ghzl/pintour/internal/db/sqlc"
	"github.com/irfan-ghzl/pintour/internal/middleware"
	"github.com/irfan-ghzl/pintour/internal/util"
	pb "github.com/irfan-ghzl/pintour/pb/pintour/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// bookingStatusToString converts proto BookingStatus to DB string
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

// stringToBookingStatus converts DB string to proto BookingStatus
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

// convertBooking converts db.Booking + participants to pb.Booking
func convertBooking(b db.Booking, participants []db.BookingParticipant, tourTitle string) *pb.Booking {
	totalPrice, _ := strconv.ParseFloat(b.TotalPrice, 64)

	var pbParticipants []*pb.BookingParticipant
	for _, p := range participants {
		dob := ""
		if p.DateOfBirth.Valid {
			dob = p.DateOfBirth.Time.Format("2006-01-02")
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
		TourTitle:       tourTitle,
		TravelDate:      b.TravelDate.Format("2006-01-02"),
		NumParticipants: b.NumParticipants,
		TotalPrice:      totalPrice,
		Status:          stringToBookingStatus(b.Status),
		Notes:           b.Notes,
		Participants:    pbParticipants,
		CreatedAt:       timestamppb.New(b.CreatedAt),
		UpdatedAt:       timestamppb.New(b.UpdatedAt),
	}
}

// CreateBooking creates a new booking
func (s *Server) CreateBooking(ctx context.Context, req *pb.CreateBookingRequest) (*pb.Booking, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}

	tp, err := s.store.GetTourPackage(ctx, req.TourPackageId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "tour package not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot get tour package: %v", err)
	}

	travelDate, err := time.Parse("2006-01-02", req.TravelDate)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid travel_date format, use YYYY-MM-DD")
	}

	price, _ := strconv.ParseFloat(tp.Price, 64)
	totalPrice := price * float64(req.NumParticipants)
	bookingCode := util.RandomBookingCode()

	var scheduleID sql.NullInt64
	if req.TourScheduleId > 0 {
		scheduleID = sql.NullInt64{Int64: req.TourScheduleId, Valid: true}
	}

	booking, err := s.store.CreateBooking(ctx, db.CreateBookingParams{
		BookingCode:     bookingCode,
		UserID:          payload.UserID,
		TourPackageID:   req.TourPackageId,
		TourScheduleID:  scheduleID,
		TravelDate:      travelDate,
		NumParticipants: req.NumParticipants,
		TotalPrice:      fmt.Sprintf("%.2f", totalPrice),
		Notes:           req.Notes,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create booking: %v", err)
	}

	var participants []db.BookingParticipant
	for _, p := range req.Participants {
		var dob sql.NullTime
		if p.DateOfBirth != "" {
			t, parseErr := time.Parse("2006-01-02", p.DateOfBirth)
			if parseErr == nil {
				dob = sql.NullTime{Time: t, Valid: true}
			}
		}
		participant, createErr := s.store.CreateBookingParticipant(ctx, db.CreateBookingParticipantParams{
			BookingID:    booking.ID,
			Name:         p.Name,
			IDCardNumber: p.IdCardNumber,
			DateOfBirth:  dob,
		})
		if createErr == nil {
			participants = append(participants, participant)
		}
	}

	return convertBooking(booking, participants, tp.Title), nil
}

// GetBooking returns a specific booking
func (s *Server) GetBooking(ctx context.Context, req *pb.GetBookingRequest) (*pb.Booking, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}

	booking, err := s.store.GetBooking(ctx, req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "booking not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot get booking: %v", err)
	}

	// Users can only see their own bookings; admins see all
	if payload.Role != "admin" && booking.UserID != payload.UserID {
		return nil, status.Errorf(codes.PermissionDenied, "access denied")
	}

	participants, _ := s.store.ListBookingParticipants(ctx, booking.ID)
	tp, _ := s.store.GetTourPackage(ctx, booking.TourPackageID)

	return convertBooking(booking, participants, tp.Title), nil
}

// ListBookings lists bookings (user sees own, admin sees all)
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
	offset := (page - 1) * limit
	statusStr := bookingStatusToString(req.Status)

	var bookings []db.Booking
	var total int64

	if payload.Role == "admin" {
		bookings, err = s.store.ListAllBookings(ctx, db.ListAllBookingsParams{
			Column1: statusStr,
			Limit:   limit,
			Offset:  offset,
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot list bookings: %v", err)
		}
		total, err = s.store.CountAllBookings(ctx, statusStr)
	} else {
		bookings, err = s.store.ListBookingsByUser(ctx, db.ListBookingsByUserParams{
			UserID:  payload.UserID,
			Column2: statusStr,
			Limit:   limit,
			Offset:  offset,
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot list bookings: %v", err)
		}
		total, err = s.store.CountBookingsByUser(ctx, db.CountBookingsByUserParams{
			UserID:  payload.UserID,
			Column2: statusStr,
		})
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot count bookings: %v", err)
	}

	var pbBookings []*pb.Booking
	for _, b := range bookings {
		participants, _ := s.store.ListBookingParticipants(ctx, b.ID)
		tp, _ := s.store.GetTourPackage(ctx, b.TourPackageID)
		pbBookings = append(pbBookings, convertBooking(b, participants, tp.Title))
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	return &pb.ListBookingsResponse{
		Bookings: pbBookings,
		Pagination: &pb.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: int32(totalPages),
		},
	}, nil
}

// CancelBooking cancels a booking
func (s *Server) CancelBooking(ctx context.Context, req *pb.CancelBookingRequest) (*pb.CancelBookingResponse, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}

	booking, err := s.store.GetBooking(ctx, req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "booking not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot get booking: %v", err)
	}

	if payload.Role != "admin" && booking.UserID != payload.UserID {
		return nil, status.Errorf(codes.PermissionDenied, "access denied")
	}

	if booking.Status == "cancelled" {
		return nil, status.Errorf(codes.FailedPrecondition, "booking is already cancelled")
	}

	updatedBooking, err := s.store.UpdateBookingStatus(ctx, db.UpdateBookingStatusParams{
		ID:     req.Id,
		Status: "cancelled",
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot cancel booking: %v", err)
	}

	participants, _ := s.store.ListBookingParticipants(ctx, updatedBooking.ID)
	tp, _ := s.store.GetTourPackage(ctx, updatedBooking.TourPackageID)

	return &pb.CancelBookingResponse{
		Message: "booking cancelled successfully",
		Booking: convertBooking(updatedBooking, participants, tp.Title),
	}, nil
}

// UpdateBookingStatus updates a booking's status (admin only)
func (s *Server) UpdateBookingStatus(ctx context.Context, req *pb.UpdateBookingStatusRequest) (*pb.Booking, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}
	if payload.Role != "admin" {
		return nil, status.Errorf(codes.PermissionDenied, "only admins can update booking status")
	}

	updatedBooking, err := s.store.UpdateBookingStatus(ctx, db.UpdateBookingStatusParams{
		ID:     req.Id,
		Status: bookingStatusToString(req.Status),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot update booking status: %v", err)
	}

	participants, _ := s.store.ListBookingParticipants(ctx, updatedBooking.ID)
	tp, _ := s.store.GetTourPackage(ctx, updatedBooking.TourPackageID)

	return convertBooking(updatedBooking, participants, tp.Title), nil
}
