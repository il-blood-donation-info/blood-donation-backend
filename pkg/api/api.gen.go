// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.15.0 DO NOT EDIT.
package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/oapi-codegen/runtime"
	strictnethttp "github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"gorm.io/gorm"
)

// Defines values for UserRole.
const (
	Admin    UserRole = "Admin"
	Reporter UserRole = "Reporter"
)

// ApiError defines model for ApiError.
type ApiError struct {
	Message string `json:"message"`
}

// SchedulePoint defines model for SchedulePoint.
type SchedulePoint struct {
	Address   string             `json:"address"`
	CloseTime string             `json:"close_time"`
	Date      openapi_types.Date `json:"date"`
	IsOpen    *bool              `json:"is_open,omitempty"`
	Name      string             `json:"name"`
	OpenTime  string             `json:"open_time"`
	StationId int64              `gorm:"primaryKey" json:"station_id"`
}

// Station defines model for Station.
type Station struct {
	Address         string             `json:"address"`
	DeletedAt       *gorm.DeletedAt    `gorm:"type:timestamp with time zone;index" json:"-"`
	Id              int64              `gorm:"primaryKey" json:"id"`
	Name            string             `json:"name"`
	StationSchedule *[]StationSchedule `json:"station_schedule,omitempty"`
}

// StationSchedule defines model for StationSchedule.
type StationSchedule struct {
	CloseTime string    `json:"close_time"`
	Date      time.Time `json:"date"`
	Id        *int64    `gorm:"primaryKey" json:"id,omitempty"`
	OpenTime  string    `json:"open_time"`

	// StationId The ID of the related station
	StationId     int64            `gorm:"index" json:"station_id"`
	StationStatus *[]StationStatus `json:"station_status,omitempty"`
}

// StationStatus defines model for StationStatus.
type StationStatus struct {
	CreatedAt time.Time `json:"created_at"`
	Id        *int64    `gorm:"primaryKey" json:"id,omitempty"`
	IsOpen    bool      `json:"is_open"`

	// StationScheduleId The ID of the related station
	StationScheduleId int64  `gorm:"index" json:"station_schedule_id"`
	UserId            *int64 `json:"user_id,omitempty"`
}

// User defines model for User.
type User struct {
	DeletedAt   *gorm.DeletedAt `gorm:"type:timestamp with time zone;index" json:"-"`
	Description string          `json:"description"`
	Email       string          `json:"email"`
	FirstName   string          `json:"first_name"`
	Id          int64           `gorm:"primaryKey" json:"id"`
	LastName    string          `json:"last_name"`
	Phone string   `json:"phone"`
	Role  UserRole `json:"role"`
}

// UserRole defines model for User.Role.
type UserRole string

// UpdateStationJSONBody defines parameters for UpdateStation.
type UpdateStationJSONBody struct {
	// IsOpen New status for station's open status
	IsOpen *bool `json:"isOpen,omitempty"`
}

// UpdateStationJSONRequestBody defines body for UpdateStation for application/json ContentType.
type UpdateStationJSONRequestBody UpdateStationJSONBody

// CreateUserJSONRequestBody defines body for CreateUser for application/json ContentType.
type CreateUserJSONRequestBody = User

// UpdateUserJSONRequestBody defines body for UpdateUser for application/json ContentType.
type UpdateUserJSONRequestBody = User

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Get the schedule for all stations with their current status
	// (GET /schedule)
	GetSchedule(w http.ResponseWriter, r *http.Request)
	// Get all stations
	// (GET /stations)
	GetStations(w http.ResponseWriter, r *http.Request)
	// Update station
	// (PUT /stations/{id})
	UpdateStation(w http.ResponseWriter, r *http.Request, id int64)
	// Get all users
	// (GET /users)
	GetUsers(w http.ResponseWriter, r *http.Request)
	// Create user
	// (POST /users)
	CreateUser(w http.ResponseWriter, r *http.Request)
	// Delete user
	// (DELETE /users/{id})
	DeleteUser(w http.ResponseWriter, r *http.Request, id int64)
	// Update user
	// (PUT /users/{id})
	UpdateUser(w http.ResponseWriter, r *http.Request, id int64)
}

// Unimplemented server implementation that returns http.StatusNotImplemented for each endpoint.

type Unimplemented struct{}

// Get the schedule for all stations with their current status
// (GET /schedule)
func (_ Unimplemented) GetSchedule(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Get all stations
// (GET /stations)
func (_ Unimplemented) GetStations(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Update station
// (PUT /stations/{id})
func (_ Unimplemented) UpdateStation(w http.ResponseWriter, r *http.Request, id int64) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Get all users
// (GET /users)
func (_ Unimplemented) GetUsers(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Create user
// (POST /users)
func (_ Unimplemented) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Delete user
// (DELETE /users/{id})
func (_ Unimplemented) DeleteUser(w http.ResponseWriter, r *http.Request, id int64) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Update user
// (PUT /users/{id})
func (_ Unimplemented) UpdateUser(w http.ResponseWriter, r *http.Request, id int64) {
	w.WriteHeader(http.StatusNotImplemented)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

// GetSchedule operation middleware
func (siw *ServerInterfaceWrapper) GetSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetSchedule(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// GetStations operation middleware
func (siw *ServerInterfaceWrapper) GetStations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetStations(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// UpdateStation operation middleware
func (siw *ServerInterfaceWrapper) UpdateStation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "id" -------------
	var id int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "id", runtime.ParamLocationPath, chi.URLParam(r, "id"), &id)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "id", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.UpdateStation(w, r, id)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// GetUsers operation middleware
func (siw *ServerInterfaceWrapper) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetUsers(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// CreateUser operation middleware
func (siw *ServerInterfaceWrapper) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.CreateUser(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// DeleteUser operation middleware
func (siw *ServerInterfaceWrapper) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "id" -------------
	var id int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "id", runtime.ParamLocationPath, chi.URLParam(r, "id"), &id)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "id", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.DeleteUser(w, r, id)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// UpdateUser operation middleware
func (siw *ServerInterfaceWrapper) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "id" -------------
	var id int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "id", runtime.ParamLocationPath, chi.URLParam(r, "id"), &id)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "id", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.UpdateUser(w, r, id)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshalingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshalingParamError) Error() string {
	return fmt.Sprintf("Error unmarshaling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshalingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{})
}

type ChiServerOptions struct {
	BaseURL          string
	BaseRouter       chi.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r chi.Router) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r chi.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options ChiServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = chi.NewRouter()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/schedule", wrapper.GetSchedule)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/stations", wrapper.GetStations)
	})
	r.Group(func(r chi.Router) {
		r.Put(options.BaseURL+"/stations/{id}", wrapper.UpdateStation)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/users", wrapper.GetUsers)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/users", wrapper.CreateUser)
	})
	r.Group(func(r chi.Router) {
		r.Delete(options.BaseURL+"/users/{id}", wrapper.DeleteUser)
	})
	r.Group(func(r chi.Router) {
		r.Put(options.BaseURL+"/users/{id}", wrapper.UpdateUser)
	})

	return r
}

type GetScheduleRequestObject struct {
}

type GetScheduleResponseObject interface {
	VisitGetScheduleResponse(w http.ResponseWriter) error
}

type GetSchedule200JSONResponse []SchedulePoint

func (response GetSchedule200JSONResponse) VisitGetScheduleResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type GetSchedule401JSONResponse ApiError

func (response GetSchedule401JSONResponse) VisitGetScheduleResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(401)

	return json.NewEncoder(w).Encode(response)
}

type GetSchedule500JSONResponse ApiError

func (response GetSchedule500JSONResponse) VisitGetScheduleResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)

	return json.NewEncoder(w).Encode(response)
}

type GetStationsRequestObject struct {
}

type GetStationsResponseObject interface {
	VisitGetStationsResponse(w http.ResponseWriter) error
}

type GetStations200JSONResponse []Station

func (response GetStations200JSONResponse) VisitGetStationsResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type GetStations401JSONResponse ApiError

func (response GetStations401JSONResponse) VisitGetStationsResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(401)

	return json.NewEncoder(w).Encode(response)
}

type GetStations500JSONResponse ApiError

func (response GetStations500JSONResponse) VisitGetStationsResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)

	return json.NewEncoder(w).Encode(response)
}

type UpdateStationRequestObject struct {
	Id   int64 `json:"id"`
	Body *UpdateStationJSONRequestBody
}

type UpdateStationResponseObject interface {
	VisitUpdateStationResponse(w http.ResponseWriter) error
}

type UpdateStation200JSONResponse Station

func (response UpdateStation200JSONResponse) VisitUpdateStationResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type UpdateStation400JSONResponse ApiError

func (response UpdateStation400JSONResponse) VisitUpdateStationResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(400)

	return json.NewEncoder(w).Encode(response)
}

type UpdateStation401JSONResponse ApiError

func (response UpdateStation401JSONResponse) VisitUpdateStationResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(401)

	return json.NewEncoder(w).Encode(response)
}

type UpdateStation403JSONResponse ApiError

func (response UpdateStation403JSONResponse) VisitUpdateStationResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(403)

	return json.NewEncoder(w).Encode(response)
}

type UpdateStation404JSONResponse ApiError

func (response UpdateStation404JSONResponse) VisitUpdateStationResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(404)

	return json.NewEncoder(w).Encode(response)
}

type UpdateStation405JSONResponse ApiError

func (response UpdateStation405JSONResponse) VisitUpdateStationResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(405)

	return json.NewEncoder(w).Encode(response)
}

type UpdateStation409JSONResponse ApiError

func (response UpdateStation409JSONResponse) VisitUpdateStationResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(409)

	return json.NewEncoder(w).Encode(response)
}

type UpdateStation500JSONResponse ApiError

func (response UpdateStation500JSONResponse) VisitUpdateStationResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)

	return json.NewEncoder(w).Encode(response)
}

type GetUsersRequestObject struct {
}

type GetUsersResponseObject interface {
	VisitGetUsersResponse(w http.ResponseWriter) error
}

type GetUsers200JSONResponse []User

func (response GetUsers200JSONResponse) VisitGetUsersResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type GetUsers401JSONResponse ApiError

func (response GetUsers401JSONResponse) VisitGetUsersResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(401)

	return json.NewEncoder(w).Encode(response)
}

type GetUsers500JSONResponse ApiError

func (response GetUsers500JSONResponse) VisitGetUsersResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)

	return json.NewEncoder(w).Encode(response)
}

type CreateUserRequestObject struct {
	Body *CreateUserJSONRequestBody
}

type CreateUserResponseObject interface {
	VisitCreateUserResponse(w http.ResponseWriter) error
}

type CreateUser201JSONResponse User

func (response CreateUser201JSONResponse) VisitCreateUserResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)

	return json.NewEncoder(w).Encode(response)
}

type CreateUser401JSONResponse ApiError

func (response CreateUser401JSONResponse) VisitCreateUserResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(401)

	return json.NewEncoder(w).Encode(response)
}

type CreateUser500JSONResponse ApiError

func (response CreateUser500JSONResponse) VisitCreateUserResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)

	return json.NewEncoder(w).Encode(response)
}

type DeleteUserRequestObject struct {
	Id int64 `json:"id"`
}

type DeleteUserResponseObject interface {
	VisitDeleteUserResponse(w http.ResponseWriter) error
}

type DeleteUser200Response struct {
}

func (response DeleteUser200Response) VisitDeleteUserResponse(w http.ResponseWriter) error {
	w.WriteHeader(200)
	return nil
}

type DeleteUser401JSONResponse ApiError

func (response DeleteUser401JSONResponse) VisitDeleteUserResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(401)

	return json.NewEncoder(w).Encode(response)
}

type DeleteUser500JSONResponse ApiError

func (response DeleteUser500JSONResponse) VisitDeleteUserResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)

	return json.NewEncoder(w).Encode(response)
}

type UpdateUserRequestObject struct {
	Id   int64 `json:"id"`
	Body *UpdateUserJSONRequestBody
}

type UpdateUserResponseObject interface {
	VisitUpdateUserResponse(w http.ResponseWriter) error
}

type UpdateUser200JSONResponse User

func (response UpdateUser200JSONResponse) VisitUpdateUserResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type UpdateUser401JSONResponse ApiError

func (response UpdateUser401JSONResponse) VisitUpdateUserResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(401)

	return json.NewEncoder(w).Encode(response)
}

type UpdateUser500JSONResponse ApiError

func (response UpdateUser500JSONResponse) VisitUpdateUserResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)

	return json.NewEncoder(w).Encode(response)
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {
	// Get the schedule for all stations with their current status
	// (GET /schedule)
	GetSchedule(ctx context.Context, request GetScheduleRequestObject) (GetScheduleResponseObject, error)
	// Get all stations
	// (GET /stations)
	GetStations(ctx context.Context, request GetStationsRequestObject) (GetStationsResponseObject, error)
	// Update station
	// (PUT /stations/{id})
	UpdateStation(ctx context.Context, request UpdateStationRequestObject) (UpdateStationResponseObject, error)
	// Get all users
	// (GET /users)
	GetUsers(ctx context.Context, request GetUsersRequestObject) (GetUsersResponseObject, error)
	// Create user
	// (POST /users)
	CreateUser(ctx context.Context, request CreateUserRequestObject) (CreateUserResponseObject, error)
	// Delete user
	// (DELETE /users/{id})
	DeleteUser(ctx context.Context, request DeleteUserRequestObject) (DeleteUserResponseObject, error)
	// Update user
	// (PUT /users/{id})
	UpdateUser(ctx context.Context, request UpdateUserRequestObject) (UpdateUserResponseObject, error)
}

type StrictHandlerFunc = strictnethttp.StrictHttpHandlerFunc
type StrictMiddlewareFunc = strictnethttp.StrictHttpMiddlewareFunc

type StrictHTTPServerOptions struct {
	RequestErrorHandlerFunc  func(w http.ResponseWriter, r *http.Request, err error)
	ResponseErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

func NewStrictHandler(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares, options: StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		},
	}}
}

func NewStrictHandlerWithOptions(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc, options StrictHTTPServerOptions) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares, options: options}
}

type strictHandler struct {
	ssi         StrictServerInterface
	middlewares []StrictMiddlewareFunc
	options     StrictHTTPServerOptions
}

// GetSchedule operation middleware
func (sh *strictHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	var request GetScheduleRequestObject

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.GetSchedule(ctx, request.(GetScheduleRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetSchedule")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(GetScheduleResponseObject); ok {
		if err := validResponse.VisitGetScheduleResponse(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("unexpected response type: %T", response))
	}
}

// GetStations operation middleware
func (sh *strictHandler) GetStations(w http.ResponseWriter, r *http.Request) {
	var request GetStationsRequestObject

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.GetStations(ctx, request.(GetStationsRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetStations")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(GetStationsResponseObject); ok {
		if err := validResponse.VisitGetStationsResponse(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("unexpected response type: %T", response))
	}
}

// UpdateStation operation middleware
func (sh *strictHandler) UpdateStation(w http.ResponseWriter, r *http.Request, id int64) {
	var request UpdateStationRequestObject

	request.Id = id

	var body UpdateStationJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sh.options.RequestErrorHandlerFunc(w, r, fmt.Errorf("can't decode JSON body: %w", err))
		return
	}
	request.Body = &body

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.UpdateStation(ctx, request.(UpdateStationRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "UpdateStation")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(UpdateStationResponseObject); ok {
		if err := validResponse.VisitUpdateStationResponse(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("unexpected response type: %T", response))
	}
}

// GetUsers operation middleware
func (sh *strictHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	var request GetUsersRequestObject

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.GetUsers(ctx, request.(GetUsersRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetUsers")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(GetUsersResponseObject); ok {
		if err := validResponse.VisitGetUsersResponse(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("unexpected response type: %T", response))
	}
}

// CreateUser operation middleware
func (sh *strictHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var request CreateUserRequestObject

	var body CreateUserJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sh.options.RequestErrorHandlerFunc(w, r, fmt.Errorf("can't decode JSON body: %w", err))
		return
	}
	request.Body = &body

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.CreateUser(ctx, request.(CreateUserRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "CreateUser")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(CreateUserResponseObject); ok {
		if err := validResponse.VisitCreateUserResponse(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("unexpected response type: %T", response))
	}
}

// DeleteUser operation middleware
func (sh *strictHandler) DeleteUser(w http.ResponseWriter, r *http.Request, id int64) {
	var request DeleteUserRequestObject

	request.Id = id

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.DeleteUser(ctx, request.(DeleteUserRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "DeleteUser")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(DeleteUserResponseObject); ok {
		if err := validResponse.VisitDeleteUserResponse(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("unexpected response type: %T", response))
	}
}

// UpdateUser operation middleware
func (sh *strictHandler) UpdateUser(w http.ResponseWriter, r *http.Request, id int64) {
	var request UpdateUserRequestObject

	request.Id = id

	var body UpdateUserJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sh.options.RequestErrorHandlerFunc(w, r, fmt.Errorf("can't decode JSON body: %w", err))
		return
	}
	request.Body = &body

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.UpdateUser(ctx, request.(UpdateUserRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "UpdateUser")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(UpdateUserResponseObject); ok {
		if err := validResponse.VisitUpdateUserResponse(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("unexpected response type: %T", response))
	}
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/+xY227jNhD9FYIt0Bf5kiYpWvXJ23SLYNtN0DRPiyBgzLHEhUSqJJXLBv73BS+6M7aT",
	"rJ0A2TdbHHKGc86ZGekez0VeCA5cKxzfYzVPISf256xgf0oppPldSFGA1AzsSg5KkQTMT31XAI6x0pLx",
	"BC+XEZbwf8kkUBx/qg0vospQXH2GucbLCJ/NU6BlBqeCcT30QSiVoFTAR4TnmVBwqVluQ1gImRONY2wf",
	"RENzSnTX0D4IGDJ1KQrgLZ9XQmRAuFnkJIdgNGbL5sEoTTQT/JLRjjnj+peDxp5xDQlIHOHbkSAFG80F",
	"hQT4CG61JCNNEpuZRMgcx7iQLCfy7gPc4WUfgpY/f4eozm1UZaK5Qie5QdjceY8DjEIGGuglsUDfjhIx",
	"+qwEH7GECwk41rKEyD33e83Nxkdu20xvmAezNzaRK03yAt0wnSLzF30RHH5nnMKtpejWcr+CJRUOytPe",
	"GDENuT3rRwkLHOMfJo0YJ16JE5/wSi7mLH84kZLcDTQXAnoFkGeteLqAfguVjR6y3iIEz9EjBTWXrHAM",
	"x/+lgI6PkFggnQKSkBENFPkdOPpG4Te8rDmiiS7Voxnidq3jR6cgPFn/Z3WIPdJIII3SX5wLKyt6X5Kv",
	"ggOlArlhc3gI2fZ9mhREbWxCyJ4rCPT6V1a6O+AEyizkhGXBlQWTSl8+WJ+3SMKMrHJcpIKHV6RwZRl4",
	"mRt8ZzRnBsd/oRBSg2yh+MD8ZQlgj+ncvx1SlbIqkG6KhzyxquILMVTK7PQYLYREOeEkYTxBV5kQFFHB",
	"LS0rzShEOEWG5mpsEsy0uSV+Z42PKuPZ6TGO8DVI5Q7fG0/H02raIgXDMd4fT8f7Jm6iU5v9Sbu3JqCH",
	"If4F2sq4MrTxkixrYnPES4FJNC+lBK6Rr8bWtbRmx9SdVfdOk3VVCK6cZn6eTm0tFFyDm21JUWRsbjdP",
	"jHyaUXvzKt+Zl4dVvi8OfPLBWB1M9x4Vy6oQ6jeCgLdzTkqdCsm+ADV+Dx+Zg6f6rRYirMrcqO7ZMDsx",
	"f8I1my7M6ZNq70pytZ0EGdOs7YAxvkl958oarvRQqwng89fFf3LP6NL2yTJAgvPCjDqt+aBLAbd8Vq8W",
	"RJIcNEjjsH+WN0PHR6aNmyem1FXTfeyre13uXSdusrR+fLhw20Hpd4LePQqC7pDA1Imfsro3+Ag3XldW",
	"gj4rPylkqnhLcb3RrNfFWr7C/aibg+UztbWRpB6W0A6pPH4p0XrH+ztx/F7IK0YpcO/1YCdePwqN3ouS",
	"V3c93InXf0CngiLjfJZl4qZO9W87ZtXhjnncqcmDIhquyHaGXNuOnVWgF5/7he03Yvtm9b0Lb9CFK7Aq",
	"wM1/fGHekoQKYPyHfaO1uwYIu7Vzt/TUNrce1U36z94WfIYSQd8qfbo86JGnLhX15Oa+aQzp5D5ahOnk",
	"1jydVo5txmbLM9uwYAVC8F9uxm+VFF00hxVlxfgeZIBbeyUMeMlyNt16OXu7jbBLwH4lM6Ygr8Os+1vM",
	"SYYoXEMmitx+0rC2OMKlzHCMU60LFU8mmTFMhdLxrwcH+9gQyjvqH3lSCUAhciVK3X5L9nRW9UeGtZur",
	"5u532kstL5ZfAwAA//+Jpw4Hgx4AAA==",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
