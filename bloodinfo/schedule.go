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
			Address:   point.StationAddress,               // This field is not provided in StationSchedulePoint
			CloseTime: point.CloseTime, // assuming you want HH:MM:SS format
			Date:      openapi_types.Date{Time: point.Date},
			IsOpen:    &point.LastStatus,
			Name:      point.StationName,
			OpenTime:  point.OpenTime, // assuming you want HH:MM:SS format
			StationId: int64(point.StationID),
		}
		schedulePoints = append(schedulePoints, schedulePoint)
	}

	return schedulePoints
}

func GetStationsFullSchedule(db *gorm.DB) ([]StationSchedulePoint, error) {
	var schedule []StationSchedulePoint

	subquery := db.Table("station_statuses").
		Select("is_open, station_schedule_id").
		Where("DATE(created_at) = DATE(?)", time.Now()).
		Order("user_id DESC, created_at DESC").
		Limit(1)

	db.Table("stations s").
		Select("s.id as station_id, " +
			"s.name as station_name, " +
			"s.address as station_address, " +
			"c.date as date, " +
			"c.open_time as open_time, " +
			"c.close_time as close_time, " +
			"COALESCE(t.is_open, false) as last_status").
		Joins("LEFT JOIN station_schedules c ON c.station_id = s.id").
		Joins("LEFT JOIN (?) as t ON t.station_schedule_id = c.id", subquery).
		Group("s.id, c.id, t.is_open").
		Order("c.date ASC").
		Scan(&schedule)

	for _, r := range schedule {
		fmt.Println(r)
	}

	return schedule, nil
}

