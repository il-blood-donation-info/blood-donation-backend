package server

import (
	"fmt"
	"github.com/il-blood-donation-info/blood-donation-backend/pkg/api"
	"github.com/oapi-codegen/runtime/types"
	"gorm.io/gorm"
	"log"
	"time"
	_ "time/tzdata"
)

type StationSchedulePoint struct {
	StationID      uint
	StationName    string
	StationAddress string
	Date           time.Time
	OpenTime       string
	CloseTime      string
	LastStatus     bool
	UserID         *int
	SchedulingUrl  string
}

const layout = "15:04"
const dateFormat = "2006-01-02"

func isOpen(point StationSchedulePoint) *bool {
	if point.UserID != nil {
		return &point.LastStatus
	}

	loc, err := time.LoadLocation("Asia/Jerusalem")
	if err != nil {
		log.Fatalf("failed to load location: %v", err)
	}
	isOpen := isOpenToday(point, loc)
	if !isOpen {
		return &isOpen
	}

	isOpen = isOpenAtThisTime(point, loc)
	return &isOpen
}

func isOpenAtThisTime(point StationSchedulePoint, loc *time.Location) bool {
	currentTime := time.Now().In(loc)
	openTime, err := time.Parse(layout, point.OpenTime)
	if err != nil {
		log.Fatalf("failed to parse open time: %v", err)
	}
	closeTime, err := time.Parse(layout, point.CloseTime)
	if err != nil {
		log.Fatalf("failed to parse close time: %v", err)
	}

	return currentTime.After(openTime) && currentTime.Before(closeTime)
}

func isOpenToday(point StationSchedulePoint, loc *time.Location) bool {
	today := time.Now().In(loc).Truncate(24 * time.Hour)
	if point.Date.After(today) {
		return false
	}

	return point.LastStatus
}

// TODO: move from here
func ConvertToSchedulePoints(points []StationSchedulePoint) []api.SchedulePoint {
	var schedulePoints []api.SchedulePoint

	for _, point := range points {
		schedulePoint := api.SchedulePoint{
			Address:       point.StationAddress, // This field is not provided in StationSchedulePoint
			CloseTime:     point.CloseTime,      // assuming you want HH:MM:SS format
			Date:          types.Date{Time: point.Date},
			IsOpen:        isOpen(point), //SchedulePoint.IsOpen is &, not showing correct value in json
			Name:          point.StationName,
			OpenTime:      point.OpenTime, // assuming you want HH:MM:SS format
			StationId:     int64(point.StationID),
			SchedulingUrl: point.SchedulingUrl,
		}
		schedulePoints = append(schedulePoints, schedulePoint)
	}

	for _, r := range schedulePoints {
		fmt.Println(r)
	}
	return schedulePoints
}

type Scheduler struct {
	SinceDate time.Time
}

type FullScheduleOptions func()

func WithSinceDate(t time.Time) func(scheduler *Scheduler) {
	return func(s *Scheduler) {
		s.SinceDate = t
	}
}
func NewScheduler(opts ...func(scheduler *Scheduler)) Scheduler {
	s := Scheduler{
		SinceDate: time.Now(),
	}
	for _, o := range opts {
		o(&s)
	}
	return s
}

type Schedule []StationSchedulePoint

func (s Schedule) FilterByDate(targetDate time.Time) Schedule {
	var filteredStations []StationSchedulePoint

	for _, station := range s {
		// Compare only the Year, Month, and Day parts of the date
		if station.Date.Year() == targetDate.Year() &&
			station.Date.Month() == targetDate.Month() &&
			station.Date.Day() == targetDate.Day() {
			filteredStations = append(filteredStations, station)
		}
	}

	return filteredStations

}

func (s *Scheduler) GetStationsFullSchedule(db *gorm.DB) (Schedule, error) {
	var schedule []StationSchedulePoint

	subquery := db.Table("station_statuses h").
		Select("h.is_open, h.user_id, h.station_schedule_id").
		//		Where(fmt.Sprintf("DATE(h.created_at) = '%s'", s.SinceDate.Format("2006-01-02"))). // Do we realy need to filter out statuses based on created date ?
		Where("h.station_schedule_id = c.id").
		Order("CASE WHEN h.user_id IS NOT NULL AND h.user_id > 0 THEN 1 ELSE 0 END DESC, h.created_at DESC").
		Limit(1)

	db.Table("stations s").
		Select("s.id as station_id, "+
			"s.name as station_name, "+
			"s.address as station_address, "+
			"c.date as date, "+
			"c.open_time as open_time, "+
			"c.close_time as close_time, "+
			"t.user_id as user_id, "+
			"COALESCE(t.is_open, true) as last_status,"+ //Don't we need to differentiate true / false and not defined ?
			"c.scheduling_url as scheduling_url").
		Joins("LEFT JOIN station_schedules c ON c.station_id = s.id").
		Joins("LEFT JOIN LATERAL (?) as t ON t.station_schedule_id = c.id", subquery).
		Where(fmt.Sprintf("date >= '%s'", s.SinceDate.Format(dateFormat))).
		Group("s.id, c.id, t.is_open, t.user_id").
		Order("c.date ASC").
		Scan(&schedule)
	//todo : Don't show past scheduled?

	return schedule, nil
}
