package repository

import (
	"context"
	"time"

	"github.com/irfan-ghzl/pintour/internal/domain/entity"
)

type ListTourPackagesFilter struct {
	Search        string
	Category      string
	DestinationID int64
	MinPrice      float64
	MaxPrice      float64
	Limit         int32
	Offset        int32
}

type TourRepository interface {
	CreatePackage(ctx context.Context, pkg entity.TourPackage) (*entity.TourPackage, error)
	GetPackage(ctx context.Context, id int64) (*entity.TourPackage, error)
	ListPackages(ctx context.Context, filter ListTourPackagesFilter) ([]entity.TourPackage, error)
	CountPackages(ctx context.Context, filter ListTourPackagesFilter) (int64, error)
	UpdatePackage(ctx context.Context, pkg entity.TourPackage) (*entity.TourPackage, error)
	DeletePackage(ctx context.Context, id int64) error

	CreateDestination(ctx context.Context, dest entity.Destination) (*entity.Destination, error)
	GetDestination(ctx context.Context, id int64) (*entity.Destination, error)
	ListDestinations(ctx context.Context, search, country string, limit, offset int32) ([]entity.Destination, error)
	CountDestinations(ctx context.Context, search, country string) (int64, error)

	CreateItinerary(ctx context.Context, tourID int64, it entity.TourItinerary) (*entity.TourItinerary, error)
	ListItineraries(ctx context.Context, tourID int64) ([]entity.TourItinerary, error)
	CreateFacility(ctx context.Context, tourID int64, name string) (*entity.TourFacility, error)
	ListFacilities(ctx context.Context, tourID int64) ([]entity.TourFacility, error)
	CreateImage(ctx context.Context, tourID int64, imageURL string) error
	ListImages(ctx context.Context, tourID int64) ([]string, error)

	CreateSchedule(ctx context.Context, sc entity.TourSchedule) (*entity.TourSchedule, error)
	GetSchedule(ctx context.Context, id int64) (*entity.TourSchedule, error)
	ListSchedules(ctx context.Context, tourID int64, startDate, endDate time.Time) ([]entity.TourSchedule, error)
	DecrementScheduleSlots(ctx context.Context, id int64, count int32) (*entity.TourSchedule, error)

	GetAverageRating(ctx context.Context, tourID int64) (float64, error)
	CountReviews(ctx context.Context, tourID int64) (int64, error)
}
