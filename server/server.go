package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/il-blood-donation-info/blood-donation-backend/pkg/api"
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
func (s StrictBloodInfoServer) GetSchedule(ctx context.Context, request api.GetScheduleRequestObject) (api.GetScheduleResponseObject, error) {
	var stationsSchedule []api.SchedulePoint
	schedule, err := s.scheduler.GetStationsFullSchedule(s.db)
	stationsSchedule = ConvertToSchedulePoints(schedule)
	if err != nil {
		return api.GetSchedule500JSONResponse{
			Message: fmt.Sprintf("error getting schedule: %w", err),
		}, nil
	}
	return api.GetSchedule200JSONResponse(stationsSchedule), nil
}

// GetStations gets all stations
func (s StrictBloodInfoServer) GetStations(ctx context.Context, request api.GetStationsRequestObject) (api.GetStationsResponseObject, error) {
	var stations []api.Station
	tx := s.db.Find(&stations)
	if tx.Error != nil {
		return api.GetStations500JSONResponse{
			Message: fmt.Sprintf("error getting stations: %v", tx.Error),
		}, nil
	}
	return api.GetStations200JSONResponse(stations), nil
}

// UpdateStation updates station
func (s StrictBloodInfoServer) UpdateStation(ctx context.Context, request api.UpdateStationRequestObject) (api.UpdateStationResponseObject, error) {
	var station api.Station
	tx := s.db.First(&station, request.Id)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			// Handle the case when no record is found
			return api.UpdateStation404JSONResponse{}, tx.Error
		} else {
			// Handle other errors
			return api.UpdateStation500JSONResponse{}, tx.Error
		}
	}

	// add a status to a schedule point
	stationStatus := api.StationStatus{
		Id:     &station.Id,
		IsOpen: request.Body.IsOpen,
	}
	tx = s.db.Create(&stationStatus)
	if tx.Error != nil {
		return api.UpdateStation500JSONResponse{}, tx.Error
	}

	return api.UpdateStation200JSONResponse{}, nil
}

// GetUsers get all users
func (s StrictBloodInfoServer) GetUsers(ctx context.Context, request api.GetUsersRequestObject) (api.GetUsersResponseObject, error) {
	var users []api.User
	tx := s.db.Find(&users)
	if tx.Error != nil {
		return api.GetUsers500JSONResponse{
			Message: fmt.Sprintf("error getting users: %v", tx.Error),
		}, nil
	}
	return api.GetUsers200JSONResponse(users), nil
}

// CreateUser creates user
func (s StrictBloodInfoServer) CreateUser(ctx context.Context, request api.CreateUserRequestObject) (api.CreateUserResponseObject, error) {
	// TODO: validate request is authorized only by admin role
	user := api.User{
		FirstName:   request.Body.FirstName,
		LastName:    request.Body.LastName,
		Email:       request.Body.Email,
		Description: request.Body.Description,
		Phone:       request.Body.Phone,
		Role:        request.Body.Role,
	}
	tx := s.db.Create(&user)

	if tx.Error != nil {
		return api.CreateUser500JSONResponse{
			Message: fmt.Sprintf("error creating user: %v", tx.Error),
		}, nil
	}

	return api.CreateUser201JSONResponse(user), nil
}

// DeleteUser deletes user
func (s StrictBloodInfoServer) DeleteUser(ctx context.Context, request api.DeleteUserRequestObject) (api.DeleteUserResponseObject, error) {
	user := api.User{
		Id: request.Id,
	}
	tx := s.db.Delete(&user)
	if tx.Error != nil {
		return api.DeleteUser500JSONResponse{
			Message: fmt.Sprintf("error deleting user: %v", tx.Error),
		}, nil
	}
	return api.DeleteUser200Response{}, nil
}

func (s StrictBloodInfoServer) UpdateUser(ctx context.Context, request api.UpdateUserRequestObject) (api.UpdateUserResponseObject, error) {
	tx := s.db.Model(&api.User{}).Where("id = ?", request.Id).Updates(api.User{
		FirstName:   request.Body.FirstName,
		LastName:    request.Body.LastName,
		Email:       request.Body.Email,
		Description: request.Body.Description,
		Phone:       request.Body.Phone,
	})
	if tx.Error != nil {
		return api.UpdateUser500JSONResponse{
			Message: fmt.Sprintf("error updating user: %v", tx.Error),
		}, nil
	}
	return api.UpdateUser200JSONResponse{}, nil
}
