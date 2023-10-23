package bloodinfo

import (
	"context"
	"fmt"
	"gorm.io/gorm"
)

// StrictBloodInfoServer implements StrictServerInterface
type StrictBloodInfoServer struct {
	db        *gorm.DB
	scheduler Scheduler
}

func NewStrictBloodInfoServer(db *gorm.DB) StrictBloodInfoServer {
	return StrictBloodInfoServer{db: db}
}

// GetSchedule gets schedule
func (s StrictBloodInfoServer) GetSchedule(ctx context.Context, request GetScheduleRequestObject) (GetScheduleResponseObject, error) {
	var stationsSchedule []SchedulePoint
	schedule, err := s.scheduler.GetStationsFullSchedule(s.db)
	stationsSchedule = ConvertToSchedulePoints(schedule)
	if err != nil {
		return GetSchedule500JSONResponse{
			Message: fmt.Sprintf("error getting schedule: %w", err),
		}, nil
	}
	return GetSchedule200JSONResponse(stationsSchedule), nil
}

// GetStations gets all stations
func (s StrictBloodInfoServer) GetStations(ctx context.Context, request GetStationsRequestObject) (GetStationsResponseObject, error) {
	var stations []Station
	tx := s.db.Find(&stations)
	if tx.Error != nil {
		return GetStations500JSONResponse{
			Message: fmt.Sprintf("error getting stations: %w", tx.Error),
		}, nil
	}
	return GetStations200JSONResponse(stations), nil
}

// UpdateStation updates station
func (s StrictBloodInfoServer) UpdateStation(ctx context.Context, request UpdateStationRequestObject) (UpdateStationResponseObject, error) {
	panic("implement me")
}

// GetUsers get all users
func (s StrictBloodInfoServer) GetUsers(ctx context.Context, request GetUsersRequestObject) (GetUsersResponseObject, error) {
	var users []User
	tx := s.db.Find(&users)
	if tx.Error != nil {
		return GetUsers500JSONResponse{
			Message: fmt.Sprintf("error getting users: %w", tx.Error),
		}, nil
	}
	return GetUsers200JSONResponse(users), nil
}

// CreateUser creates user
func (s StrictBloodInfoServer) CreateUser(ctx context.Context, request CreateUserRequestObject) (CreateUserResponseObject, error) {
	// TODO: validate request is authorized only by admin role
	user := User{
		FirstName:   request.Body.FirstName,
		LastName:    request.Body.LastName,
		Email:       request.Body.Email,
		Description: request.Body.Description,
		Phone:       request.Body.Phone,
		Role:        request.Body.Role,
	}
	tx := s.db.Create(&user)

	if tx.Error != nil {
		return CreateUser500JSONResponse{
			Message: fmt.Sprintf("error creating user: %w", tx.Error),
		}, nil
	}

	return CreateUser201JSONResponse(user), nil
}

// DeleteUser deletes user
func (s StrictBloodInfoServer) DeleteUser(ctx context.Context, request DeleteUserRequestObject) (DeleteUserResponseObject, error) {
	user := User{
		Id: request.Id,
	}
	tx := s.db.Delete(&user)
	if tx.Error != nil {
		return DeleteUser500JSONResponse{
			Message: fmt.Sprintf("error deleting user: %w", tx.Error),
		}, nil
	}
	return DeleteUser200Response{}, nil
}

func (s StrictBloodInfoServer) UpdateUser(ctx context.Context, request UpdateUserRequestObject) (UpdateUserResponseObject, error) {
	tx := s.db.Model(&User{}).Where("id = ?", request.Id).Updates(User{
		FirstName:   request.Body.FirstName,
		LastName:    request.Body.LastName,
		Email:       request.Body.Email,
		Description: request.Body.Description,
		Phone:       request.Body.Phone,
	})
	if tx.Error != nil {
		return UpdateUser500JSONResponse{
			Message: fmt.Sprintf("error updating user: %w", tx.Error),
		}, nil
	}
	return UpdateUser200JSONResponse{}, nil
}
