package review

import "time"

type CreateReviewInput struct {
	UserID        int64
	TourPackageID int64
	BookingID     int64
	Rating        int32
	Comment       string
}

type ListReviewsInput struct {
	TourPackageID int64
	Page          int32
	Limit         int32
}

type ListReviewsOutput struct {
	Reviews       []ReviewWithUser
	Total         int64
	TotalPages    int32
	Page          int32
	Limit         int32
	AverageRating float64
}

type ReviewWithUser struct {
	ID            int64
	UserID        int64
	UserName      string
	UserAvatar    string
	TourPackageID int64
	BookingID     int64
	Rating        int32
	Comment       string
	CreatedAt     time.Time
}
