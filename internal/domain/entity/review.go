package entity

import "time"

type Review struct {
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
