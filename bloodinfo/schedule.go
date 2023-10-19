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

func GetStationsFullSchedule(db *gorm.DB) ([]StationSchedulePoint, error) {
	var schedule []StationSchedulePoint

	subquery := db.Table("station_statuses h").
		Select("h.is_open, h.station_schedule_id").
		Where("DATE(h.created_at) = CURRENT_DATE").
		Where("h.station_schedule_id = c.id").
		Order("CASE WHEN h.user_id IS NOT NULL AND h.user_id > 0 THEN 1 ELSE 0 END DESC, h.created_at DESC").
		Limit(1)

	db.Table("stations s").
		Select("s.id as station_id, " +
			"s.name as station_name, " +
			"s.address as station_address, " +
			"c.date as date, " +
			"c.open_time as open_time, " +
			"c.close_time as close_time, " +
			"t.is_open as last_status").
		Joins("LEFT JOIN station_schedules c ON c.station_id = s.id").
		Joins("LEFT JOIN LATERAL (?) as t ON t.station_schedule_id = c.id", subquery).
		Group("s.id, c.id, t.is_open").
		Order("s.id, c.date ASC").
		Scan(&schedule)
	//todo : Don't show past scheduled?

	for _, r := range schedule {
		fmt.Println(r)
	}

	return schedule, nil
}

