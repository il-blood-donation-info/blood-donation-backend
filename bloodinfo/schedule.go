package bloodinfo

import (
	"fmt"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"gorm.io/gorm"
	"time"
)

type StationSchedulePoint struct {
	StationID      uint
	StationName    string
	StationAddress string
	Date           time.Time
	OpenTime       string
	CloseTime      string
	LastStatus     bool
}

// TODO: move from here
func ConvertToSchedulePoints(points []StationSchedulePoint) []SchedulePoint {
	var schedulePoints []SchedulePoint

	for _, point := range points {
		schedulePoint := SchedulePoint{
			Address:   point.StationAddress, // This field is not provided in StationSchedulePoint
			CloseTime: point.CloseTime,      // assuming you want HH:MM:SS format
			Date:      openapi_types.Date{Time: point.Date},
			IsOpen:    &point.LastStatus, //SchedulePoint.IsOpen is &, not showing correct value in json
			Name:      point.StationName,
			OpenTime:  point.OpenTime, // assuming you want HH:MM:SS format
			StationId: int64(point.StationID),
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
		//func() time.Time {
		//return time.Date(2023, 10, 18, 0, 0, 0, 0, nil)
		//}
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
		Select("h.is_open, h.station_schedule_id").
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
			"COALESCE(t.is_open, true) as last_status"). //Dont we need to differenciate true / false and not defined ?
		Joins("LEFT JOIN station_schedules c ON c.station_id = s.id").
		Joins("LEFT JOIN LATERAL (?) as t ON t.station_schedule_id = c.id", subquery).
		Where(fmt.Sprintf("date >= '%s'", s.SinceDate.Format("2006-01-02"))).
		Group("s.id, c.id, t.is_open").
		Order("c.date ASC").
		Scan(&schedule)
	//todo : Don't show past scheduled?

	for _, r := range schedule {
		fmt.Println(r)
	}

	return schedule, nil
}
