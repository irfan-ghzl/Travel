package tour

import (
	"github.com/irfan-ghzl/pintour/internal/domain/entity"
)

type ListPackagesInput struct {
	Search        string
	Category      entity.TourCategory
	DestinationID int64
	MinPrice      float64
	MaxPrice      float64
	Page          int32
	Limit         int32
}

type ListPackagesOutput struct {
	Packages   []entity.TourPackage
	Total      int64
	TotalPages int32
	Page       int32
	Limit      int32
}

type CreatePackageInput struct {
	Title           string
	Description     string
	DestinationID   int64
	Price           float64
	DurationDays    int32
	MaxParticipants int32
	MinParticipants int32
	Category        entity.TourCategory
	ImageURL        string
}

type UpdatePackageInput struct {
	ID              int64
	Title           *string
	Description     *string
	Price           *float64
	DurationDays    *int32
	MaxParticipants *int32
	MinParticipants *int32
	Category        *entity.TourCategory
	ImageURL        *string
	IsActive        *bool
}

type ListDestinationsInput struct {
	Search  string
	Country string
	Page    int32
	Limit   int32
}

type ListDestinationsOutput struct {
	Destinations []entity.Destination
	Total        int64
	TotalPages   int32
	Page         int32
	Limit        int32
}

type ListSchedulesInput struct {
	TourPackageID int64
	StartDate     string
	EndDate       string
}

type CreateScheduleInput struct {
	TourPackageID  int64
	StartDate      string
	EndDate        string
	AvailableSlots int32
}
