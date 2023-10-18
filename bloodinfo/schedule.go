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
	ScheduleID     uint
	Date           time.Time
	OpenTime       time.Time
	CloseTime      time.Time
	LastStatusID   uint
	LastStatus     *bool
}

// TODO: move from here
func ConvertToSchedulePoints(points []StationSchedulePoint) []SchedulePoint {
	var schedulePoints []SchedulePoint

	for _, point := range points {
		schedulePoint := SchedulePoint{
			Address:   point.StationAddress,               // This field is not provided in StationSchedulePoint
			CloseTime: point.CloseTime.Format("15:04:05"), // assuming you want HH:MM:SS format
			Date:      openapi_types.Date{Time: point.Date},
			IsOpen:    point.LastStatus,
			Name:      point.StationName,
			OpenTime:  point.OpenTime.Format("15:04:05"), // assuming you want HH:MM:SS format
			StationId: int64(point.StationID),
		}
		schedulePoints = append(schedulePoints, schedulePoint)
	}

	return schedulePoints
}

func GetStationsFullSchedule(db *gorm.DB) ([]StationSchedulePoint, error) {
	var schedule []StationSchedulePoint

	today := time.Now().Format("2006-01-02")

	subQuery := db.Select("MAX(id) as id").
		Table("station_statuses").
		Order(fmt.Sprintf("CASE WHEN DATE(when_set_by_user) = %s THEN 1 ELSE 0 END DESC, id DESC", today)).
		Group("station_schedule_id")

	db.Table("stations").
		Select("stations.id as station_id, stations.name as station_name, stations.address as station_address, station_schedules.id as schedule_id, station_schedules.date, station_schedules.open_time, station_schedules.close_time, station_statuses.id as last_status_id, station_statuses.is_open as last_status").
		Joins("JOIN station_schedules ON stations.id = station_schedules.station_id").
		Joins("LEFT JOIN (?) AS sub_status ON station_schedules.id = sub_status.station_schedule_id", subQuery).
		Joins("LEFT JOIN station_statuses ON sub_status.id = station_statuses.id").
		Order("station_schedules.date").
		Scan(&schedule)

	return schedule, nil
}
