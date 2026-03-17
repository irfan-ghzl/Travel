package booking

import (
	"context"
	"database/sql"
	"time"

	"github.com/irfan-ghzl/pintour/internal/domain/entity"
	"github.com/irfan-ghzl/pintour/internal/domain/repository"
	"github.com/irfan-ghzl/pintour/common/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	bookingRepo repository.BookingRepository
	tourRepo    repository.TourRepository
}

func NewService(bookingRepo repository.BookingRepository, tourRepo repository.TourRepository) *Service {
	return &Service{bookingRepo: bookingRepo, tourRepo: tourRepo}
}

func (s *Service) CreateBooking(ctx context.Context, input CreateBookingInput) (*entity.Booking, error) {
	pkg, err := s.tourRepo.GetPackage(ctx, input.TourPackageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "tour package not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot get tour package: %v", err)
	}

	travelDate, err := time.Parse("2006-01-02", input.TravelDate)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid travel_date format, use YYYY-MM-DD")
	}

	totalPrice := pkg.Price * float64(input.NumParticipants)
	bookingCode := util.RandomBookingCode()

	var scheduleID *int64
	if input.TourScheduleID > 0 {
		id := input.TourScheduleID
		scheduleID = &id
	}

	booking, err := s.bookingRepo.Create(ctx, repository.CreateBookingParams{
		BookingCode:     bookingCode,
		UserID:          input.UserID,
		TourPackageID:   input.TourPackageID,
		TourScheduleID:  scheduleID,
		TravelDate:      travelDate,
		NumParticipants: input.NumParticipants,
		TotalPrice:      totalPrice,
		Notes:           input.Notes,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create booking: %v", err)
	}

	for _, p := range input.Participants {
		var dob *time.Time
		if p.DateOfBirth != "" {
			t, parseErr := time.Parse("2006-01-02", p.DateOfBirth)
			if parseErr == nil {
				dob = &t
			}
		}
		participant, createErr := s.bookingRepo.CreateParticipant(ctx, entity.BookingParticipant{
			BookingID:    booking.ID,
			Name:         p.Name,
			IDCardNumber: p.IDCardNumber,
			DateOfBirth:  dob,
		})
		if createErr == nil {
			booking.Participants = append(booking.Participants, *participant)
		}
	}

	booking.TourTitle = pkg.Title
	return booking, nil
}

func (s *Service) GetBooking(ctx context.Context, id int64, userID int64, role string) (*entity.Booking, error) {
	booking, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "booking not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot get booking: %v", err)
	}

	if role != "admin" && booking.UserID != userID {
		return nil, status.Errorf(codes.PermissionDenied, "access denied")
	}

	participants, _ := s.bookingRepo.ListParticipants(ctx, booking.ID)
	booking.Participants = participants

	pkg, _ := s.tourRepo.GetPackage(ctx, booking.TourPackageID)
	if pkg != nil {
		booking.TourTitle = pkg.Title
	}

	return booking, nil
}

func (s *Service) ListBookings(ctx context.Context, input ListBookingsInput) (*ListBookingsOutput, error) {
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

	var bookings []entity.Booking
	var total int64
	var err error

	if input.Role == "admin" {
		bookings, err = s.bookingRepo.ListAll(ctx, statusStr, limit, offset)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot list bookings: %v", err)
		}
		total, err = s.bookingRepo.CountAll(ctx, statusStr)
	} else {
		bookings, err = s.bookingRepo.ListByUser(ctx, input.UserID, statusStr, limit, offset)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot list bookings: %v", err)
		}
		total, err = s.bookingRepo.CountByUser(ctx, input.UserID, statusStr)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot count bookings: %v", err)
	}

	for i := range bookings {
		participants, _ := s.bookingRepo.ListParticipants(ctx, bookings[i].ID)
		bookings[i].Participants = participants
		pkg, _ := s.tourRepo.GetPackage(ctx, bookings[i].TourPackageID)
		if pkg != nil {
			bookings[i].TourTitle = pkg.Title
		}
	}

	totalPages := int32((total + int64(limit) - 1) / int64(limit))
	return &ListBookingsOutput{
		Bookings:   bookings,
		Total:      total,
		TotalPages: totalPages,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (s *Service) CancelBooking(ctx context.Context, id int64, userID int64, role string) (*entity.Booking, error) {
	booking, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "booking not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot get booking: %v", err)
	}

	if role != "admin" && booking.UserID != userID {
		return nil, status.Errorf(codes.PermissionDenied, "access denied")
	}

	if booking.Status == entity.BookingStatusCancelled {
		return nil, status.Errorf(codes.FailedPrecondition, "booking is already cancelled")
	}

	updated, err := s.bookingRepo.UpdateStatus(ctx, id, entity.BookingStatusCancelled)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot cancel booking: %v", err)
	}

	participants, _ := s.bookingRepo.ListParticipants(ctx, updated.ID)
	updated.Participants = participants
	pkg, _ := s.tourRepo.GetPackage(ctx, updated.TourPackageID)
	if pkg != nil {
		updated.TourTitle = pkg.Title
	}

	return updated, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id int64, newStatus entity.BookingStatus) (*entity.Booking, error) {
	updated, err := s.bookingRepo.UpdateStatus(ctx, id, newStatus)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot update booking status: %v", err)
	}

	participants, _ := s.bookingRepo.ListParticipants(ctx, updated.ID)
	updated.Participants = participants
	pkg, _ := s.tourRepo.GetPackage(ctx, updated.TourPackageID)
	if pkg != nil {
		updated.TourTitle = pkg.Title
	}

	return updated, nil
}
