package entity

import "time"

type TourCategory string

const (
	TourCategoryAdventure TourCategory = "adventure"
	TourCategoryCultural  TourCategory = "cultural"
	TourCategoryBeach     TourCategory = "beach"
	TourCategoryCity      TourCategory = "city"
	TourCategoryNature    TourCategory = "nature"
	TourCategoryReligious TourCategory = "religious"
	TourCategoryHoneymoon TourCategory = "honeymoon"
	TourCategoryFamily    TourCategory = "family"
)

type Destination struct {
	ID          int64
	Name        string
	Country     string
	City        string
	Description string
	ImageURL    string
	CreatedAt   time.Time
}

type TourPackage struct {
	ID              int64
	Title           string
	Description     string
	DestinationID   int64
	Destination     *Destination
	Price           float64
	DurationDays    int32
	MaxParticipants int32
	MinParticipants int32
	Category        TourCategory
	ImageURL        string
	IsActive        bool
	AverageRating   float64
	ReviewCount     int64
	Itineraries     []TourItinerary
	Facilities      []TourFacility
	Images          []string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type TourItinerary struct {
	ID          int64
	DayNumber   int32
	Title       string
	Description string
}

type TourFacility struct {
	ID   int64
	Name string
}

type TourSchedule struct {
	ID             int64
	TourPackageID  int64
	StartDate      time.Time
	EndDate        time.Time
	AvailableSlots int32
	Status         string
	CreatedAt      time.Time
}
