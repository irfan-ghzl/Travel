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

type tourRepository struct {
	q db.Querier
}

func NewTourRepository(q db.Querier) repository.TourRepository {
	return &tourRepository{q: q}
}

func (r *tourRepository) CreatePackage(ctx context.Context, pkg entity.TourPackage) (*entity.TourPackage, error) {
	result, err := r.q.CreateTourPackage(ctx, db.CreateTourPackageParams{
		Title:           pkg.Title,
		Description:     pkg.Description,
		DestinationID:   pkg.DestinationID,
		Price:           fmt.Sprintf("%.2f", pkg.Price),
		DurationDays:    pkg.DurationDays,
		MaxParticipants: pkg.MaxParticipants,
		MinParticipants: pkg.MinParticipants,
		Category:        string(pkg.Category),
		ImageUrl:        pkg.ImageURL,
	})
	if err != nil {
		return nil, err
	}
	return toEntityTourPackage(result), nil
}

func (r *tourRepository) GetPackage(ctx context.Context, id int64) (*entity.TourPackage, error) {
	result, err := r.q.GetTourPackage(ctx, id)
	if err != nil {
		return nil, err
	}
	pkg := toEntityTourPackage(result)

	dest, err := r.q.GetDestination(ctx, result.DestinationID)
	if err == nil {
		pkg.Destination = toEntityDestination(dest)
	}

	itineraries, _ := r.q.ListTourItineraries(ctx, id)
	for _, it := range itineraries {
		pkg.Itineraries = append(pkg.Itineraries, entity.TourItinerary{
			ID:          it.ID,
			DayNumber:   it.DayNumber,
			Title:       it.Title,
			Description: it.Description,
		})
	}

	facilities, _ := r.q.ListTourFacilities(ctx, id)
	for _, f := range facilities {
		pkg.Facilities = append(pkg.Facilities, entity.TourFacility{
			ID:   f.ID,
			Name: f.Name,
		})
	}

	images, _ := r.q.ListTourImages(ctx, id)
	for _, img := range images {
		pkg.Images = append(pkg.Images, img.ImageUrl)
	}

	avgStr, _ := r.q.GetAverageRating(ctx, id)
	pkg.AverageRating, _ = strconv.ParseFloat(avgStr, 64)
	pkg.ReviewCount, _ = r.q.CountReviewsByTour(ctx, id)

	return pkg, nil
}

func (r *tourRepository) ListPackages(ctx context.Context, filter repository.ListTourPackagesFilter) ([]entity.TourPackage, error) {
	minPrice := "0"
	maxPrice := "0"
	if filter.MinPrice > 0 {
		minPrice = fmt.Sprintf("%.2f", filter.MinPrice)
	}
	if filter.MaxPrice > 0 {
		maxPrice = fmt.Sprintf("%.2f", filter.MaxPrice)
	}

	results, err := r.q.ListTourPackages(ctx, db.ListTourPackagesParams{
		Column1: filter.Search,
		Column2: filter.Category,
		Column3: filter.DestinationID,
		Column4: minPrice,
		Column5: maxPrice,
		Limit:   filter.Limit,
		Offset:  filter.Offset,
	})
	if err != nil {
		return nil, err
	}

	var packages []entity.TourPackage
	for _, tp := range results {
		pkg := toEntityTourPackage(tp)
		dest, err := r.q.GetDestination(ctx, tp.DestinationID)
		if err == nil {
			pkg.Destination = toEntityDestination(dest)
		}
		avgStr, _ := r.q.GetAverageRating(ctx, tp.ID)
		pkg.AverageRating, _ = strconv.ParseFloat(avgStr, 64)
		pkg.ReviewCount, _ = r.q.CountReviewsByTour(ctx, tp.ID)
		packages = append(packages, *pkg)
	}
	return packages, nil
}

func (r *tourRepository) CountPackages(ctx context.Context, filter repository.ListTourPackagesFilter) (int64, error) {
	minPrice := "0"
	maxPrice := "0"
	if filter.MinPrice > 0 {
		minPrice = fmt.Sprintf("%.2f", filter.MinPrice)
	}
	if filter.MaxPrice > 0 {
		maxPrice = fmt.Sprintf("%.2f", filter.MaxPrice)
	}

	return r.q.CountTourPackages(ctx, db.CountTourPackagesParams{
		Column1: filter.Search,
		Column2: filter.Category,
		Column3: filter.DestinationID,
		Column4: minPrice,
		Column5: maxPrice,
	})
}

func (r *tourRepository) UpdatePackage(ctx context.Context, pkg entity.TourPackage) (*entity.TourPackage, error) {
	priceStr := fmt.Sprintf("%.2f", pkg.Price)

	result, err := r.q.UpdateTourPackage(ctx, db.UpdateTourPackageParams{
		ID:              pkg.ID,
		Title:           sql.NullString{String: pkg.Title, Valid: true},
		Description:     sql.NullString{String: pkg.Description, Valid: true},
		Price:           sql.NullString{String: priceStr, Valid: true},
		DurationDays:    sql.NullInt32{Int32: pkg.DurationDays, Valid: true},
		MaxParticipants: sql.NullInt32{Int32: pkg.MaxParticipants, Valid: true},
		MinParticipants: sql.NullInt32{Int32: pkg.MinParticipants, Valid: true},
		Category:        sql.NullString{String: string(pkg.Category), Valid: true},
		ImageUrl:        sql.NullString{String: pkg.ImageURL, Valid: true},
		IsActive:        sql.NullBool{Bool: pkg.IsActive, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return toEntityTourPackage(result), nil
}

func (r *tourRepository) DeletePackage(ctx context.Context, id int64) error {
	return r.q.DeleteTourPackage(ctx, id)
}

func (r *tourRepository) CreateDestination(ctx context.Context, dest entity.Destination) (*entity.Destination, error) {
	result, err := r.q.CreateDestination(ctx, db.CreateDestinationParams{
		Name:        dest.Name,
		Country:     dest.Country,
		City:        dest.City,
		Description: dest.Description,
		ImageUrl:    dest.ImageURL,
	})
	if err != nil {
		return nil, err
	}
	return toEntityDestination(result), nil
}

func (r *tourRepository) GetDestination(ctx context.Context, id int64) (*entity.Destination, error) {
	result, err := r.q.GetDestination(ctx, id)
	if err != nil {
		return nil, err
	}
	return toEntityDestination(result), nil
}

func (r *tourRepository) ListDestinations(ctx context.Context, search, country string, limit, offset int32) ([]entity.Destination, error) {
	results, err := r.q.ListDestinations(ctx, db.ListDestinationsParams{
		Column1: search,
		Column2: country,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, err
	}
	var dests []entity.Destination
	for _, d := range results {
		dests = append(dests, *toEntityDestination(d))
	}
	return dests, nil
}

func (r *tourRepository) CountDestinations(ctx context.Context, search, country string) (int64, error) {
	return r.q.CountDestinations(ctx, db.CountDestinationsParams{
		Column1: search,
		Column2: country,
	})
}

func (r *tourRepository) CreateItinerary(ctx context.Context, tourID int64, it entity.TourItinerary) (*entity.TourItinerary, error) {
	result, err := r.q.CreateTourItinerary(ctx, db.CreateTourItineraryParams{
		TourPackageID: tourID,
		DayNumber:     it.DayNumber,
		Title:         it.Title,
		Description:   it.Description,
	})
	if err != nil {
		return nil, err
	}
	return &entity.TourItinerary{
		ID:          result.ID,
		DayNumber:   result.DayNumber,
		Title:       result.Title,
		Description: result.Description,
	}, nil
}

func (r *tourRepository) ListItineraries(ctx context.Context, tourID int64) ([]entity.TourItinerary, error) {
	results, err := r.q.ListTourItineraries(ctx, tourID)
	if err != nil {
		return nil, err
	}
	var items []entity.TourItinerary
	for _, it := range results {
		items = append(items, entity.TourItinerary{
			ID:          it.ID,
			DayNumber:   it.DayNumber,
			Title:       it.Title,
			Description: it.Description,
		})
	}
	return items, nil
}

func (r *tourRepository) CreateFacility(ctx context.Context, tourID int64, name string) (*entity.TourFacility, error) {
	result, err := r.q.CreateTourFacility(ctx, db.CreateTourFacilityParams{
		TourPackageID: tourID,
		Name:          name,
	})
	if err != nil {
		return nil, err
	}
	return &entity.TourFacility{ID: result.ID, Name: result.Name}, nil
}

func (r *tourRepository) ListFacilities(ctx context.Context, tourID int64) ([]entity.TourFacility, error) {
	results, err := r.q.ListTourFacilities(ctx, tourID)
	if err != nil {
		return nil, err
	}
	var items []entity.TourFacility
	for _, f := range results {
		items = append(items, entity.TourFacility{ID: f.ID, Name: f.Name})
	}
	return items, nil
}

func (r *tourRepository) CreateImage(ctx context.Context, tourID int64, imageURL string) error {
	_, err := r.q.CreateTourImage(ctx, db.CreateTourImageParams{
		TourPackageID: tourID,
		ImageUrl:      imageURL,
	})
	return err
}

func (r *tourRepository) ListImages(ctx context.Context, tourID int64) ([]string, error) {
	results, err := r.q.ListTourImages(ctx, tourID)
	if err != nil {
		return nil, err
	}
	var images []string
	for _, img := range results {
		images = append(images, img.ImageUrl)
	}
	return images, nil
}

func (r *tourRepository) CreateSchedule(ctx context.Context, sc entity.TourSchedule) (*entity.TourSchedule, error) {
	result, err := r.q.CreateTourSchedule(ctx, db.CreateTourScheduleParams{
		TourPackageID:  sc.TourPackageID,
		StartDate:      sc.StartDate,
		EndDate:        sc.EndDate,
		AvailableSlots: sc.AvailableSlots,
	})
	if err != nil {
		return nil, err
	}
	return toEntitySchedule(result), nil
}

func (r *tourRepository) GetSchedule(ctx context.Context, id int64) (*entity.TourSchedule, error) {
	result, err := r.q.GetTourSchedule(ctx, id)
	if err != nil {
		return nil, err
	}
	return toEntitySchedule(result), nil
}

func (r *tourRepository) ListSchedules(ctx context.Context, tourID int64, startDate, endDate time.Time) ([]entity.TourSchedule, error) {
	results, err := r.q.ListTourSchedules(ctx, db.ListTourSchedulesParams{
		TourPackageID: tourID,
		Column2:       startDate,
		Column3:       endDate,
	})
	if err != nil {
		return nil, err
	}
	var schedules []entity.TourSchedule
	for _, sc := range results {
		schedules = append(schedules, *toEntitySchedule(sc))
	}
	return schedules, nil
}

func (r *tourRepository) DecrementScheduleSlots(ctx context.Context, id int64, count int32) (*entity.TourSchedule, error) {
	result, err := r.q.UpdateTourScheduleSlots(ctx, db.UpdateTourScheduleSlotsParams{
		ID:             id,
		AvailableSlots: count,
	})
	if err != nil {
		return nil, err
	}
	return toEntitySchedule(result), nil
}

func (r *tourRepository) GetAverageRating(ctx context.Context, tourID int64) (float64, error) {
	avgStr, err := r.q.GetAverageRating(ctx, tourID)
	if err != nil {
		return 0, err
	}
	avg, _ := strconv.ParseFloat(avgStr, 64)
	return avg, nil
}

func (r *tourRepository) CountReviews(ctx context.Context, tourID int64) (int64, error) {
	return r.q.CountReviewsByTour(ctx, tourID)
}

func toEntityTourPackage(tp db.TourPackage) *entity.TourPackage {
	price, _ := strconv.ParseFloat(tp.Price, 64)
	return &entity.TourPackage{
		ID:              tp.ID,
		Title:           tp.Title,
		Description:     tp.Description,
		DestinationID:   tp.DestinationID,
		Price:           price,
		DurationDays:    tp.DurationDays,
		MaxParticipants: tp.MaxParticipants,
		MinParticipants: tp.MinParticipants,
		Category:        entity.TourCategory(tp.Category),
		ImageURL:        tp.ImageUrl,
		IsActive:        tp.IsActive,
		CreatedAt:       tp.CreatedAt,
		UpdatedAt:       tp.UpdatedAt,
	}
}

func toEntityDestination(d db.Destination) *entity.Destination {
	return &entity.Destination{
		ID:          d.ID,
		Name:        d.Name,
		Country:     d.Country,
		City:        d.City,
		Description: d.Description,
		ImageURL:    d.ImageUrl,
		CreatedAt:   d.CreatedAt,
	}
}

func toEntitySchedule(sc db.TourSchedule) *entity.TourSchedule {
	return &entity.TourSchedule{
		ID:             sc.ID,
		TourPackageID:  sc.TourPackageID,
		StartDate:      sc.StartDate,
		EndDate:        sc.EndDate,
		AvailableSlots: sc.AvailableSlots,
		Status:         sc.Status,
		CreatedAt:      sc.CreatedAt,
	}
}
