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
	OpenTime       time.Time
	CloseTime      time.Time
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

	//today := time.Now().Format("2006-01-02")

	// First subquery: Identifying statuses set by the user today
	//todayStatuses := db.Select("id").
	//	Table("station_statuses").
	//	Where("DATE(when_set_by_user) = ?", today)

	// Second subquery: Get the latest status if no status exists from today
	//latestStatuses := db.Select("MAX(id) as id").
	//	Table("station_statuses").
	//	Where("id NOT IN (?)", todayStatuses).
	//	Group("station_schedule_id")
	//
	//// Main query
	//db.Table("stations").
	//	//Select("stations.id as station_id, stations.name as station_name, stations.address as station_address, station_schedules.id as schedule_id, station_schedules.date, station_schedules.open_time, station_schedules.close_time, station_statuses.id as last_status_id, station_statuses.is_open as last_status").
	//	Select("stations.id as station_id, stations.name as station_name, stations.address as station_address, station_schedules.id as schedule_id, station_schedules.date, station_schedules.open_time, station_schedules.close_time").
	//	Joins("JOIN station_schedules ON stations.id = station_schedules.station_id").
	//	//Joins("LEFT JOIN (?) AS sub_status ON station_schedules.id = sub_status.station_schedule_id", latestStatuses).
	//	//Joins("LEFT JOIN station_statuses ON sub_status.id = station_statuses.id").
	//	Order("station_schedules.date").
	//	Scan(&schedule)
	//
	//StationID      uint
	//StationName    string
	//StationAddress string
	//Date           time.Time
	//OpenTime       time.Time
	//CloseTime      time.Time
	//LastStatus     *bool

	subquery := db.Table("station_statuses").
		Select("is_open, station_schedule_id").
		Where("DATE(created_at) = DATE(?)", time.Now()).
		Order("user_id DESC, created_at DESC").
		Limit(1)

	db.Table("stations s").
		Select("s.id as station_id, s.name as station_name, s.address as station_address, c.date as date, c.open_time as open_time, c.close_time as close_time, t.is_open as last_status").
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
