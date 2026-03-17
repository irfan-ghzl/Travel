package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	db "github.com/irfan-ghzl/pintour/db/sqlc"
	"github.com/irfan-ghzl/pintour/internal/domain/entity"
	"github.com/irfan-ghzl/pintour/internal/domain/repository"
)

type bookingRepository struct {
	q db.Querier
}

func NewBookingRepository(q db.Querier) repository.BookingRepository {
	return &bookingRepository{q: q}
}

func (r *bookingRepository) Create(ctx context.Context, params repository.CreateBookingParams) (*entity.Booking, error) {
	travelDate, _ := params.TravelDate.(time.Time)

	var scheduleID sql.NullInt64
	if params.TourScheduleID != nil {
		scheduleID = sql.NullInt64{Int64: *params.TourScheduleID, Valid: true}
	}

	result, err := r.q.CreateBooking(ctx, db.CreateBookingParams{
		BookingCode:     params.BookingCode,
		UserID:          params.UserID,
		TourPackageID:   params.TourPackageID,
		TourScheduleID:  scheduleID,
		TravelDate:      travelDate,
		NumParticipants: params.NumParticipants,
		TotalPrice:      fmt.Sprintf("%.2f", params.TotalPrice),
		Notes:           params.Notes,
	})
	if err != nil {
		return nil, err
	}
	return toEntityBooking(result), nil
}

func (r *bookingRepository) GetByID(ctx context.Context, id int64) (*entity.Booking, error) {
	result, err := r.q.GetBooking(ctx, id)
	if err != nil {
		return nil, err
	}
	return toEntityBooking(result), nil
}

func (r *bookingRepository) GetByCode(ctx context.Context, code string) (*entity.Booking, error) {
	result, err := r.q.GetBookingByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return toEntityBooking(result), nil
}

func (r *bookingRepository) ListByUser(ctx context.Context, userID int64, status string, limit, offset int32) ([]entity.Booking, error) {
	results, err := r.q.ListBookingsByUser(ctx, db.ListBookingsByUserParams{
		UserID:  userID,
		Column2: status,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, err
	}
	var bookings []entity.Booking
	for _, b := range results {
		bookings = append(bookings, *toEntityBooking(b))
	}
	return bookings, nil
}

func (r *bookingRepository) CountByUser(ctx context.Context, userID int64, status string) (int64, error) {
	return r.q.CountBookingsByUser(ctx, db.CountBookingsByUserParams{
		UserID:  userID,
		Column2: status,
	})
}

func (r *bookingRepository) ListAll(ctx context.Context, status string, limit, offset int32) ([]entity.Booking, error) {
	results, err := r.q.ListAllBookings(ctx, db.ListAllBookingsParams{
		Column1: status,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, err
	}
	var bookings []entity.Booking
	for _, b := range results {
		bookings = append(bookings, *toEntityBooking(b))
	}
	return bookings, nil
}

func (r *bookingRepository) CountAll(ctx context.Context, status string) (int64, error) {
	return r.q.CountAllBookings(ctx, status)
}

func (r *bookingRepository) UpdateStatus(ctx context.Context, id int64, status entity.BookingStatus) (*entity.Booking, error) {
	result, err := r.q.UpdateBookingStatus(ctx, db.UpdateBookingStatusParams{
		ID:     id,
		Status: string(status),
	})
	if err != nil {
		return nil, err
	}
	return toEntityBooking(result), nil
}

func (r *bookingRepository) CreateParticipant(ctx context.Context, p entity.BookingParticipant) (*entity.BookingParticipant, error) {
	var dob sql.NullTime
	if p.DateOfBirth != nil {
		dob = sql.NullTime{Time: *p.DateOfBirth, Valid: true}
	}

	result, err := r.q.CreateBookingParticipant(ctx, db.CreateBookingParticipantParams{
		BookingID:    p.BookingID,
		Name:         p.Name,
		IDCardNumber: p.IDCardNumber,
		DateOfBirth:  dob,
	})
	if err != nil {
		return nil, err
	}

	ep := &entity.BookingParticipant{
		ID:           result.ID,
		BookingID:    result.BookingID,
		Name:         result.Name,
		IDCardNumber: result.IDCardNumber,
	}
	if result.DateOfBirth.Valid {
		ep.DateOfBirth = &result.DateOfBirth.Time
	}
	return ep, nil
}

func (r *bookingRepository) ListParticipants(ctx context.Context, bookingID int64) ([]entity.BookingParticipant, error) {
	results, err := r.q.ListBookingParticipants(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	var participants []entity.BookingParticipant
	for _, p := range results {
		ep := entity.BookingParticipant{
			ID:           p.ID,
			BookingID:    p.BookingID,
			Name:         p.Name,
			IDCardNumber: p.IDCardNumber,
		}
		if p.DateOfBirth.Valid {
			ep.DateOfBirth = &p.DateOfBirth.Time
		}
		participants = append(participants, ep)
	}
	return participants, nil
}

func toEntityBooking(b db.Booking) *entity.Booking {
	totalPrice, _ := strconv.ParseFloat(b.TotalPrice, 64)
	booking := &entity.Booking{
		ID:              b.ID,
		BookingCode:     b.BookingCode,
		UserID:          b.UserID,
		TourPackageID:   b.TourPackageID,
		TravelDate:      b.TravelDate,
		NumParticipants: b.NumParticipants,
		TotalPrice:      totalPrice,
		Status:          entity.BookingStatus(b.Status),
		Notes:           b.Notes,
		CreatedAt:       b.CreatedAt,
		UpdatedAt:       b.UpdatedAt,
	}
	if b.TourScheduleID.Valid {
		id := b.TourScheduleID.Int64
		booking.TourScheduleID = &id
	}
	return booking
}
