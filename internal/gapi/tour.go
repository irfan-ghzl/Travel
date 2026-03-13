package gapi

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	db "github.com/irfan-ghzl/pintour/internal/db/sqlc"
	"github.com/irfan-ghzl/pintour/internal/middleware"
	pb "github.com/irfan-ghzl/pintour/pb/pintour/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// tourCategoryToString converts proto TourCategory to DB string
func tourCategoryToString(cat pb.TourCategory) string {
	switch cat {
	case pb.TourCategory_TOUR_CATEGORY_ADVENTURE:
		return "adventure"
	case pb.TourCategory_TOUR_CATEGORY_CULTURAL:
		return "cultural"
	case pb.TourCategory_TOUR_CATEGORY_BEACH:
		return "beach"
	case pb.TourCategory_TOUR_CATEGORY_CITY:
		return "city"
	case pb.TourCategory_TOUR_CATEGORY_NATURE:
		return "nature"
	case pb.TourCategory_TOUR_CATEGORY_RELIGIOUS:
		return "religious"
	case pb.TourCategory_TOUR_CATEGORY_HONEYMOON:
		return "honeymoon"
	case pb.TourCategory_TOUR_CATEGORY_FAMILY:
		return "family"
	default:
		return ""
	}
}

// stringToTourCategory converts DB string to proto TourCategory
func stringToTourCategory(s string) pb.TourCategory {
	switch s {
	case "adventure":
		return pb.TourCategory_TOUR_CATEGORY_ADVENTURE
	case "cultural":
		return pb.TourCategory_TOUR_CATEGORY_CULTURAL
	case "beach":
		return pb.TourCategory_TOUR_CATEGORY_BEACH
	case "city":
		return pb.TourCategory_TOUR_CATEGORY_CITY
	case "nature":
		return pb.TourCategory_TOUR_CATEGORY_NATURE
	case "religious":
		return pb.TourCategory_TOUR_CATEGORY_RELIGIOUS
	case "honeymoon":
		return pb.TourCategory_TOUR_CATEGORY_HONEYMOON
	case "family":
		return pb.TourCategory_TOUR_CATEGORY_FAMILY
	default:
		return pb.TourCategory_TOUR_CATEGORY_UNSPECIFIED
	}
}

// convertTourPackage converts a db.TourPackage to pb.TourPackage
func convertTourPackage(tp db.TourPackage, dest *db.Destination, itineraries []db.TourItinerary, facilities []db.TourFacility, images []db.TourImage, avgRating float64, reviewCount int64) *pb.TourPackage {
	price, _ := strconv.ParseFloat(tp.Price, 64)

	var pbDest *pb.Destination
	if dest != nil {
		pbDest = convertDestination(*dest)
	}

	var pbItineraries []*pb.TourItinerary
	for _, it := range itineraries {
		pbItineraries = append(pbItineraries, &pb.TourItinerary{
			Id:          it.ID,
			DayNumber:   it.DayNumber,
			Title:       it.Title,
			Description: it.Description,
		})
	}

	var pbFacilities []*pb.TourFacility
	for _, f := range facilities {
		pbFacilities = append(pbFacilities, &pb.TourFacility{
			Id:   f.ID,
			Name: f.Name,
		})
	}

	var imageURLs []string
	for _, img := range images {
		imageURLs = append(imageURLs, img.ImageUrl)
	}

	return &pb.TourPackage{
		Id:              tp.ID,
		Title:           tp.Title,
		Description:     tp.Description,
		Destination:     pbDest,
		Price:           price,
		DurationDays:    tp.DurationDays,
		MaxParticipants: tp.MaxParticipants,
		MinParticipants: tp.MinParticipants,
		Category:        stringToTourCategory(tp.Category),
		ImageUrl:        tp.ImageUrl,
		IsActive:        tp.IsActive,
		AverageRating:   avgRating,
		ReviewCount:     int32(reviewCount),
		Itineraries:     pbItineraries,
		Facilities:      pbFacilities,
		Images:          imageURLs,
		CreatedAt:       timestamppb.New(tp.CreatedAt),
		UpdatedAt:       timestamppb.New(tp.UpdatedAt),
	}
}

// convertDestination converts a db.Destination to pb.Destination
func convertDestination(dest db.Destination) *pb.Destination {
	return &pb.Destination{
		Id:          dest.ID,
		Name:        dest.Name,
		Country:     dest.Country,
		City:        dest.City,
		Description: dest.Description,
		ImageUrl:    dest.ImageUrl,
		CreatedAt:   timestamppb.New(dest.CreatedAt),
	}
}

// ListTourPackages lists tour packages with filtering and pagination
func (s *Server) ListTourPackages(ctx context.Context, req *pb.ListTourPackagesRequest) (*pb.ListTourPackagesResponse, error) {
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

	categoryStr := tourCategoryToString(req.Category)
	minPrice := fmt.Sprintf("%.2f", req.MinPrice)
	maxPrice := fmt.Sprintf("%.2f", req.MaxPrice)
	if req.MinPrice == 0 {
		minPrice = "0"
	}
	if req.MaxPrice == 0 {
		maxPrice = "0"
	}

	packages, err := s.store.ListTourPackages(ctx, db.ListTourPackagesParams{
		Column1: req.Search,
		Column2: categoryStr,
		Column3: req.DestinationId,
		Column4: minPrice,
		Column5: maxPrice,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list tour packages: %v", err)
	}

	total, err := s.store.CountTourPackages(ctx, db.CountTourPackagesParams{
		Column1: req.Search,
		Column2: categoryStr,
		Column3: req.DestinationId,
		Column4: minPrice,
		Column5: maxPrice,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot count tour packages: %v", err)
	}

	var pbPackages []*pb.TourPackage
	for _, tp := range packages {
		dest, _ := s.store.GetDestination(ctx, tp.DestinationID)
		avgRatingStr, _ := s.store.GetAverageRating(ctx, tp.ID)
		avgRating, _ := strconv.ParseFloat(avgRatingStr, 64)
		reviewCount, _ := s.store.CountReviewsByTour(ctx, tp.ID)
		pbPackages = append(pbPackages, convertTourPackage(tp, &dest, nil, nil, nil, avgRating, reviewCount))
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	return &pb.ListTourPackagesResponse{
		Packages: pbPackages,
		Pagination: &pb.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: int32(totalPages),
		},
	}, nil
}

// GetTourPackage returns a specific tour package with full details
func (s *Server) GetTourPackage(ctx context.Context, req *pb.GetTourPackageRequest) (*pb.TourPackage, error) {
	tp, err := s.store.GetTourPackage(ctx, req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "tour package not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot get tour package: %v", err)
	}

	dest, _ := s.store.GetDestination(ctx, tp.DestinationID)
	itineraries, _ := s.store.ListTourItineraries(ctx, tp.ID)
	facilities, _ := s.store.ListTourFacilities(ctx, tp.ID)
	images, _ := s.store.ListTourImages(ctx, tp.ID)
	avgRatingStr, _ := s.store.GetAverageRating(ctx, tp.ID)
	avgRating, _ := strconv.ParseFloat(avgRatingStr, 64)
	reviewCount, _ := s.store.CountReviewsByTour(ctx, tp.ID)

	return convertTourPackage(tp, &dest, itineraries, facilities, images, avgRating, reviewCount), nil
}

// CreateTourPackage creates a new tour package (admin only)
func (s *Server) CreateTourPackage(ctx context.Context, req *pb.CreateTourPackageRequest) (*pb.TourPackage, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}
	if payload.Role != "admin" {
		return nil, status.Errorf(codes.PermissionDenied, "only admins can create tour packages")
	}

	tp, err := s.store.CreateTourPackage(ctx, db.CreateTourPackageParams{
		Title:           req.Title,
		Description:     req.Description,
		DestinationID:   req.DestinationId,
		Price:           fmt.Sprintf("%.2f", req.Price),
		DurationDays:    req.DurationDays,
		MaxParticipants: req.MaxParticipants,
		MinParticipants: req.MinParticipants,
		Category:        tourCategoryToString(req.Category),
		ImageUrl:        req.ImageUrl,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create tour package: %v", err)
	}

	dest, _ := s.store.GetDestination(ctx, tp.DestinationID)
	return convertTourPackage(tp, &dest, nil, nil, nil, 0, 0), nil
}

// UpdateTourPackage updates a tour package (admin only)
func (s *Server) UpdateTourPackage(ctx context.Context, req *pb.UpdateTourPackageRequest) (*pb.TourPackage, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}
	if payload.Role != "admin" {
		return nil, status.Errorf(codes.PermissionDenied, "only admins can update tour packages")
	}

	params := db.UpdateTourPackageParams{ID: req.Id}
	if req.Title != "" {
		params.Title = sql.NullString{String: req.Title, Valid: true}
	}
	if req.Description != "" {
		params.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.Price > 0 {
		params.Price = sql.NullString{String: fmt.Sprintf("%.2f", req.Price), Valid: true}
	}
	if req.DurationDays > 0 {
		params.DurationDays = sql.NullInt32{Int32: req.DurationDays, Valid: true}
	}
	if req.MaxParticipants > 0 {
		params.MaxParticipants = sql.NullInt32{Int32: req.MaxParticipants, Valid: true}
	}
	if req.MinParticipants > 0 {
		params.MinParticipants = sql.NullInt32{Int32: req.MinParticipants, Valid: true}
	}
	if req.Category != pb.TourCategory_TOUR_CATEGORY_UNSPECIFIED {
		params.Category = sql.NullString{String: tourCategoryToString(req.Category), Valid: true}
	}
	if req.ImageUrl != "" {
		params.ImageUrl = sql.NullString{String: req.ImageUrl, Valid: true}
	}
	params.IsActive = sql.NullBool{Bool: req.IsActive, Valid: true}

	tp, err := s.store.UpdateTourPackage(ctx, params)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot update tour package: %v", err)
	}

	dest, _ := s.store.GetDestination(ctx, tp.DestinationID)
	return convertTourPackage(tp, &dest, nil, nil, nil, 0, 0), nil
}

// DeleteTourPackage soft-deletes a tour package (admin only)
func (s *Server) DeleteTourPackage(ctx context.Context, req *pb.DeleteTourPackageRequest) (*pb.DeleteTourPackageResponse, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}
	if payload.Role != "admin" {
		return nil, status.Errorf(codes.PermissionDenied, "only admins can delete tour packages")
	}

	err = s.store.DeleteTourPackage(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot delete tour package: %v", err)
	}

	return &pb.DeleteTourPackageResponse{Message: "tour package deleted successfully"}, nil
}

// ListDestinations lists all destinations with filtering
func (s *Server) ListDestinations(ctx context.Context, req *pb.ListDestinationsRequest) (*pb.ListDestinationsResponse, error) {
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

	destinations, err := s.store.ListDestinations(ctx, db.ListDestinationsParams{
		Column1: req.Search,
		Column2: req.Country,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list destinations: %v", err)
	}

	total, err := s.store.CountDestinations(ctx, db.CountDestinationsParams{
		Column1: req.Search,
		Column2: req.Country,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot count destinations: %v", err)
	}

	var pbDests []*pb.Destination
	for _, d := range destinations {
		pbDests = append(pbDests, convertDestination(d))
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	return &pb.ListDestinationsResponse{
		Destinations: pbDests,
		Pagination: &pb.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: int32(totalPages),
		},
	}, nil
}

// CreateDestination creates a new destination (admin only)
func (s *Server) CreateDestination(ctx context.Context, req *pb.CreateDestinationRequest) (*pb.Destination, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}
	if payload.Role != "admin" {
		return nil, status.Errorf(codes.PermissionDenied, "only admins can create destinations")
	}

	dest, err := s.store.CreateDestination(ctx, db.CreateDestinationParams{
		Name:        req.Name,
		Country:     req.Country,
		City:        req.City,
		Description: req.Description,
		ImageUrl:    req.ImageUrl,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create destination: %v", err)
	}

	return convertDestination(dest), nil
}

// ListTourSchedules lists schedules for a tour package
func (s *Server) ListTourSchedules(ctx context.Context, req *pb.ListTourSchedulesRequest) (*pb.ListTourSchedulesResponse, error) {
	var startDate, endDate time.Time
	if req.StartDate != "" {
		startDate, _ = time.Parse("2006-01-02", req.StartDate)
	}
	if req.EndDate != "" {
		endDate, _ = time.Parse("2006-01-02", req.EndDate)
	}

	schedules, err := s.store.ListTourSchedules(ctx, db.ListTourSchedulesParams{
		TourPackageID: req.TourPackageId,
		Column2:       startDate,
		Column3:       endDate,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list tour schedules: %v", err)
	}

	var pbSchedules []*pb.TourSchedule
	for _, sc := range schedules {
		pbSchedules = append(pbSchedules, convertTourSchedule(sc))
	}

	return &pb.ListTourSchedulesResponse{Schedules: pbSchedules}, nil
}

// CreateTourSchedule creates a new tour schedule (admin only)
func (s *Server) CreateTourSchedule(ctx context.Context, req *pb.CreateTourScheduleRequest) (*pb.TourSchedule, error) {
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return nil, err
	}
	if payload.Role != "admin" {
		return nil, status.Errorf(codes.PermissionDenied, "only admins can create tour schedules")
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid start_date format, use YYYY-MM-DD")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid end_date format, use YYYY-MM-DD")
	}

	schedule, err := s.store.CreateTourSchedule(ctx, db.CreateTourScheduleParams{
		TourPackageID:  req.TourPackageId,
		StartDate:      startDate,
		EndDate:        endDate,
		AvailableSlots: req.AvailableSlots,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create tour schedule: %v", err)
	}

	return convertTourSchedule(schedule), nil
}

// convertTourSchedule converts a db.TourSchedule to pb.TourSchedule
func convertTourSchedule(sc db.TourSchedule) *pb.TourSchedule {
	return &pb.TourSchedule{
		Id:             sc.ID,
		TourPackageId:  sc.TourPackageID,
		StartDate:      sc.StartDate.Format("2006-01-02"),
		EndDate:        sc.EndDate.Format("2006-01-02"),
		AvailableSlots: sc.AvailableSlots,
		Status:         sc.Status,
		CreatedAt:      timestamppb.New(sc.CreatedAt),
	}
}
