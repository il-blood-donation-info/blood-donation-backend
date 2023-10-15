package bloodinfo

import (
	"context"
	"fmt"
	"gorm.io/gorm"
)

// StrictBloodInfoServer implements StrictServerInterface
type StrictBloodInfoServer struct {
	db *gorm.DB
}

func NewStrictBloodInfoServer(db *gorm.DB) StrictBloodInfoServer {
	return StrictBloodInfoServer{db: db}
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
	panic("implement me")
}

// CreateUser creates user
func (s StrictBloodInfoServer) CreateUser(ctx context.Context, request CreateUserRequestObject) (CreateUserResponseObject, error) {
	panic("implement me")
}

// DeleteUser deletes user
func (s StrictBloodInfoServer) DeleteUser(ctx context.Context, request DeleteUserRequestObject) (DeleteUserResponseObject, error) {
	panic("implement me")
}

func (s StrictBloodInfoServer) UpdateUser(ctx context.Context, request UpdateUserRequestObject) (UpdateUserResponseObject, error) {
	panic("implement me")
}
