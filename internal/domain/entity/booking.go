package entity

import "time"

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusCancelled BookingStatus = "cancelled"
	BookingStatusCompleted BookingStatus = "completed"
)

type Booking struct {
	ID              int64
	BookingCode     string
	UserID          int64
	TourPackageID   int64
	TourScheduleID  *int64
	TravelDate      time.Time
	NumParticipants int32
	TotalPrice      float64
	Status          BookingStatus
	Notes           string
	TourTitle       string
	Participants    []BookingParticipant
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type BookingParticipant struct {
	ID           int64
	BookingID    int64
	Name         string
	IDCardNumber string
	DateOfBirth  *time.Time
}
