package tour

import (
	"context"
	"database/sql"
	"time"

	"github.com/irfan-ghzl/pintour/internal/domain/entity"
	"github.com/irfan-ghzl/pintour/internal/domain/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	tourRepo repository.TourRepository
}

func NewService(tourRepo repository.TourRepository) *Service {
	return &Service{tourRepo: tourRepo}
}

func (s *Service) ListPackages(ctx context.Context, input ListPackagesInput) (*ListPackagesOutput, error) {
	page := input.Page
	if page < 1 {
		page = 1
	}
	limit := input.Limit
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	filter := repository.ListTourPackagesFilter{
		Search:        input.Search,
		Category:      string(input.Category),
		DestinationID: input.DestinationID,
		MinPrice:      input.MinPrice,
		MaxPrice:      input.MaxPrice,
		Limit:         limit,
		Offset:        offset,
	}

	packages, err := s.tourRepo.ListPackages(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list tour packages: %v", err)
	}

	total, err := s.tourRepo.CountPackages(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot count tour packages: %v", err)
	}

	totalPages := int32((total + int64(limit) - 1) / int64(limit))
	return &ListPackagesOutput{
		Packages:   packages,
		Total:      total,
		TotalPages: totalPages,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (s *Service) GetPackage(ctx context.Context, id int64) (*entity.TourPackage, error) {
	pkg, err := s.tourRepo.GetPackage(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "tour package not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot get tour package: %v", err)
	}
	return pkg, nil
}

func (s *Service) CreatePackage(ctx context.Context, input CreatePackageInput) (*entity.TourPackage, error) {
	pkg, err := s.tourRepo.CreatePackage(ctx, entity.TourPackage{
		Title:           input.Title,
		Description:     input.Description,
		DestinationID:   input.DestinationID,
		Price:           input.Price,
		DurationDays:    input.DurationDays,
		MaxParticipants: input.MaxParticipants,
		MinParticipants: input.MinParticipants,
		Category:        input.Category,
		ImageURL:        input.ImageURL,
		IsActive:        true,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create tour package: %v", err)
	}

	dest, err := s.tourRepo.GetDestination(ctx, pkg.DestinationID)
	if err == nil {
		pkg.Destination = dest
	}

	return pkg, nil
}

func (s *Service) UpdatePackage(ctx context.Context, input UpdatePackageInput) (*entity.TourPackage, error) {
	existing, err := s.tourRepo.GetPackage(ctx, input.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "tour package not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot get tour package: %v", err)
	}

	if input.Title != nil {
		existing.Title = *input.Title
	}
	if input.Description != nil {
		existing.Description = *input.Description
	}
	if input.Price != nil {
		existing.Price = *input.Price
	}
	if input.DurationDays != nil {
		existing.DurationDays = *input.DurationDays
	}
	if input.MaxParticipants != nil {
		existing.MaxParticipants = *input.MaxParticipants
	}
	if input.MinParticipants != nil {
		existing.MinParticipants = *input.MinParticipants
	}
	if input.Category != nil {
		existing.Category = *input.Category
	}
	if input.ImageURL != nil {
		existing.ImageURL = *input.ImageURL
	}
	if input.IsActive != nil {
		existing.IsActive = *input.IsActive
	}

	pkg, err := s.tourRepo.UpdatePackage(ctx, *existing)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot update tour package: %v", err)
	}

	dest, err := s.tourRepo.GetDestination(ctx, pkg.DestinationID)
	if err == nil {
		pkg.Destination = dest
	}

	return pkg, nil
}

func (s *Service) DeletePackage(ctx context.Context, id int64) error {
	if err := s.tourRepo.DeletePackage(ctx, id); err != nil {
		return status.Errorf(codes.Internal, "cannot delete tour package: %v", err)
	}
	return nil
}

func (s *Service) ListDestinations(ctx context.Context, input ListDestinationsInput) (*ListDestinationsOutput, error) {
	page := input.Page
	if page < 1 {
		page = 1
	}
	limit := input.Limit
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	destinations, err := s.tourRepo.ListDestinations(ctx, input.Search, input.Country, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list destinations: %v", err)
	}

	total, err := s.tourRepo.CountDestinations(ctx, input.Search, input.Country)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot count destinations: %v", err)
	}

	totalPages := int32((total + int64(limit) - 1) / int64(limit))
	return &ListDestinationsOutput{
		Destinations: destinations,
		Total:        total,
		TotalPages:   totalPages,
		Page:         page,
		Limit:        limit,
	}, nil
}

func (s *Service) CreateDestination(ctx context.Context, dest entity.Destination) (*entity.Destination, error) {
	result, err := s.tourRepo.CreateDestination(ctx, dest)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create destination: %v", err)
	}
	return result, nil
}

func (s *Service) ListSchedules(ctx context.Context, input ListSchedulesInput) ([]entity.TourSchedule, error) {
	var startDate, endDate time.Time
	if input.StartDate != "" {
		startDate, _ = time.Parse("2006-01-02", input.StartDate)
	}
	if input.EndDate != "" {
		endDate, _ = time.Parse("2006-01-02", input.EndDate)
	}

	schedules, err := s.tourRepo.ListSchedules(ctx, input.TourPackageID, startDate, endDate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot list schedules: %v", err)
	}
	return schedules, nil
}

func (s *Service) CreateSchedule(ctx context.Context, input CreateScheduleInput) (*entity.TourSchedule, error) {
	startDate, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid start_date format, use YYYY-MM-DD")
	}
	endDate, err := time.Parse("2006-01-02", input.EndDate)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid end_date format, use YYYY-MM-DD")
	}

	schedule, err := s.tourRepo.CreateSchedule(ctx, entity.TourSchedule{
		TourPackageID:  input.TourPackageID,
		StartDate:      startDate,
		EndDate:        endDate,
		AvailableSlots: input.AvailableSlots,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create schedule: %v", err)
	}
	return schedule, nil
}
