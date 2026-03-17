package grpc

import (
"context"

apptour "github.com/irfan-ghzl/pintour/internal/application/tour"
"github.com/irfan-ghzl/pintour/internal/domain/entity"
"github.com/irfan-ghzl/pintour/common/interface/middleware"
pb "github.com/irfan-ghzl/pintour/pb/pintour/v1"
"google.golang.org/protobuf/types/known/timestamppb"
)

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

out, err := s.tourService.ListPackages(ctx, apptour.ListPackagesInput{
Search:        req.Search,
Category:      entity.TourCategory(tourCategoryToString(req.Category)),
DestinationID: req.DestinationId,
MinPrice:      req.MinPrice,
MaxPrice:      req.MaxPrice,
Page:          page,
Limit:         limit,
})
if err != nil {
return nil, err
}

var pbPackages []*pb.TourPackage
for i := range out.Packages {
pbPackages = append(pbPackages, convertTourPackage(&out.Packages[i]))
}
return &pb.ListTourPackagesResponse{
Packages: pbPackages,
Pagination: &pb.PaginationResponse{
Page:       out.Page,
Limit:      out.Limit,
Total:      out.Total,
TotalPages: out.TotalPages,
},
}, nil
}

func (s *Server) GetTourPackage(ctx context.Context, req *pb.GetTourPackageRequest) (*pb.TourPackage, error) {
pkg, err := s.tourService.GetPackage(ctx, req.Id)
if err != nil {
return nil, err
}
return convertTourPackage(pkg), nil
}

func (s *Server) CreateTourPackage(ctx context.Context, req *pb.CreateTourPackageRequest) (*pb.TourPackage, error) {
_, err := middleware.GetPayload(ctx)
if err != nil {
return nil, err
}

pkg, err := s.tourService.CreatePackage(ctx, apptour.CreatePackageInput{
Title:           req.Title,
Description:     req.Description,
DestinationID:   req.DestinationId,
Price:           req.Price,
DurationDays:    req.DurationDays,
MaxParticipants: req.MaxParticipants,
MinParticipants: req.MinParticipants,
Category:        entity.TourCategory(tourCategoryToString(req.Category)),
ImageURL:        req.ImageUrl,
})
if err != nil {
return nil, err
}
return convertTourPackage(pkg), nil
}

func (s *Server) UpdateTourPackage(ctx context.Context, req *pb.UpdateTourPackageRequest) (*pb.TourPackage, error) {
_, err := middleware.GetPayload(ctx)
if err != nil {
return nil, err
}

input := apptour.UpdatePackageInput{ID: req.Id}
if req.Title != "" {
input.Title = &req.Title
}
if req.Description != "" {
input.Description = &req.Description
}
if req.Price > 0 {
input.Price = &req.Price
}
if req.DurationDays > 0 {
input.DurationDays = &req.DurationDays
}
if req.MaxParticipants > 0 {
input.MaxParticipants = &req.MaxParticipants
}
if req.MinParticipants > 0 {
input.MinParticipants = &req.MinParticipants
}
if req.Category != pb.TourCategory_TOUR_CATEGORY_UNSPECIFIED {
cat := entity.TourCategory(tourCategoryToString(req.Category))
input.Category = &cat
}
if req.ImageUrl != "" {
input.ImageURL = &req.ImageUrl
}
isActive := req.IsActive
input.IsActive = &isActive

pkg, err := s.tourService.UpdatePackage(ctx, input)
if err != nil {
return nil, err
}
return convertTourPackage(pkg), nil
}

func (s *Server) DeleteTourPackage(ctx context.Context, req *pb.DeleteTourPackageRequest) (*pb.DeleteTourPackageResponse, error) {
_, err := middleware.GetPayload(ctx)
if err != nil {
return nil, err
}

if err := s.tourService.DeletePackage(ctx, req.Id); err != nil {
return nil, err
}
return &pb.DeleteTourPackageResponse{Message: "tour package deleted successfully"}, nil
}

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

out, err := s.tourService.ListDestinations(ctx, apptour.ListDestinationsInput{
Search:  req.Search,
Country: req.Country,
Page:    page,
Limit:   limit,
})
if err != nil {
return nil, err
}

var pbDests []*pb.Destination
for _, d := range out.Destinations {
pbDests = append(pbDests, convertDestination(&d))
}
return &pb.ListDestinationsResponse{
Destinations: pbDests,
Pagination: &pb.PaginationResponse{
Page:       out.Page,
Limit:      out.Limit,
Total:      out.Total,
TotalPages: out.TotalPages,
},
}, nil
}

func (s *Server) CreateDestination(ctx context.Context, req *pb.CreateDestinationRequest) (*pb.Destination, error) {
_, err := middleware.GetPayload(ctx)
if err != nil {
return nil, err
}

dest, err := s.tourService.CreateDestination(ctx, entity.Destination{
Name:        req.Name,
Country:     req.Country,
City:        req.City,
Description: req.Description,
ImageURL:    req.ImageUrl,
})
if err != nil {
return nil, err
}
return convertDestination(dest), nil
}

func (s *Server) ListTourSchedules(ctx context.Context, req *pb.ListTourSchedulesRequest) (*pb.ListTourSchedulesResponse, error) {
schedules, err := s.tourService.ListSchedules(ctx, apptour.ListSchedulesInput{
TourPackageID: req.TourPackageId,
StartDate:     req.StartDate,
EndDate:       req.EndDate,
})
if err != nil {
return nil, err
}

var pbSchedules []*pb.TourSchedule
for _, sc := range schedules {
pbSchedules = append(pbSchedules, convertTourSchedule(&sc))
}
return &pb.ListTourSchedulesResponse{Schedules: pbSchedules}, nil
}

func (s *Server) CreateTourSchedule(ctx context.Context, req *pb.CreateTourScheduleRequest) (*pb.TourSchedule, error) {
_, err := middleware.GetPayload(ctx)
if err != nil {
return nil, err
}

sc, err := s.tourService.CreateSchedule(ctx, apptour.CreateScheduleInput{
TourPackageID:  req.TourPackageId,
StartDate:      req.StartDate,
EndDate:        req.EndDate,
AvailableSlots: req.AvailableSlots,
})
if err != nil {
return nil, err
}
return convertTourSchedule(sc), nil
}

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

func convertTourPackage(pkg *entity.TourPackage) *pb.TourPackage {
pbPkg := &pb.TourPackage{
Id:              pkg.ID,
Title:           pkg.Title,
Description:     pkg.Description,
Price:           pkg.Price,
DurationDays:    pkg.DurationDays,
MaxParticipants: pkg.MaxParticipants,
MinParticipants: pkg.MinParticipants,
Category:        stringToTourCategory(string(pkg.Category)),
ImageUrl:        pkg.ImageURL,
IsActive:        pkg.IsActive,
AverageRating:   pkg.AverageRating,
ReviewCount:     int32(pkg.ReviewCount),
Images:          pkg.Images,
CreatedAt:       timestamppb.New(pkg.CreatedAt),
UpdatedAt:       timestamppb.New(pkg.UpdatedAt),
}

if pkg.Destination != nil {
pbPkg.Destination = convertDestination(pkg.Destination)
}

for _, it := range pkg.Itineraries {
pbPkg.Itineraries = append(pbPkg.Itineraries, &pb.TourItinerary{
Id:          it.ID,
DayNumber:   it.DayNumber,
Title:       it.Title,
Description: it.Description,
})
}

for _, f := range pkg.Facilities {
pbPkg.Facilities = append(pbPkg.Facilities, &pb.TourFacility{
Id:   f.ID,
Name: f.Name,
})
}

return pbPkg
}

func convertDestination(d *entity.Destination) *pb.Destination {
return &pb.Destination{
Id:          d.ID,
Name:        d.Name,
Country:     d.Country,
City:        d.City,
Description: d.Description,
ImageUrl:    d.ImageURL,
CreatedAt:   timestamppb.New(d.CreatedAt),
}
}

func convertTourSchedule(sc *entity.TourSchedule) *pb.TourSchedule {
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
