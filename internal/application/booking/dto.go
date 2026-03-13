package booking

import (
	"github.com/irfan-ghzl/pintour/internal/domain/entity"
)

type ParticipantInput struct {
	Name         string
	IDCardNumber string
	DateOfBirth  string
}

type CreateBookingInput struct {
	UserID          int64
	TourPackageID   int64
	TourScheduleID  int64
	TravelDate      string
	NumParticipants int32
	Notes           string
	Participants    []ParticipantInput
}

type ListBookingsInput struct {
	UserID int64
	Role   string
	Status entity.BookingStatus
	Page   int32
	Limit  int32
}

type ListBookingsOutput struct {
	Bookings   []entity.Booking
	Total      int64
	TotalPages int32
	Page       int32
	Limit      int32
}
