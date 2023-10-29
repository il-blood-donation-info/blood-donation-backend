package scraper

import (
	"fmt"
	"github.com/il-blood-donation-info/blood-donation-backend/pkg/api"
	"gorm.io/gorm"
	"log"
	"strings"
	"time"
)

type ScheduleDataWriter struct {
	DB        *gorm.DB
	SinceTime time.Time
}

func (p ScheduleDataWriter) SaveData(donationDetails []DonationDetail) error {
	tx := p.DB.Begin()
	if tx.Error != nil {
		return tx.Error // Check for an error when starting the transaction
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // Rollback the transaction in case of a panic
		}
	}()

	log.Printf("Saving data... %d \n\n", len(donationDetails))
	err := p.processDonationDetails(tx, donationDetails)
	if err != nil {
		tx.Rollback() // Rollback the transaction if an error occurs
		return err
	}

	tx.Commit() // Commit the transaction if no errors occurred
	return nil
}

func (p ScheduleDataWriter) processDonationDetails(tx *gorm.DB, donationDetails []DonationDetail) error {
	schedulesIdsMap := make(map[int64]bool)
	for _, donation := range donationDetails {
		stationName := strings.TrimSpace(donation.Name)
		stationAddress := strings.TrimSpace(fmt.Sprintf("%s %s %s", strings.TrimSpace(donation.City), strings.TrimSpace(donation.NumHouse), strings.TrimSpace(donation.Street)))

		var station = api.Station{}
		if err := tx.FirstOrInit(&station, api.Station{Name: stationName}).Error; err != nil {
			log.Printf("Error while fetching or initializing station: %v", err)
			return err
		}
		station.Address = stationAddress

		if !p.isDatePassed(donation.DateDonation) {
			var schedule = api.StationSchedule{}
			if err := tx.FirstOrInit(&schedule, api.StationSchedule{
				StationId: station.Id,
				Date:      donation.DateDonation,
				OpenTime:  donation.FromHour,
				CloseTime: donation.ToHour,
			}).Error; err != nil {
				log.Printf("Error while fetching or initializing schedule: %v", err)
				return err
			}

			if schedule.Id == nil && p.isScheduleToday(schedule) {
				schedule.StationStatus = &[]api.StationStatus{{IsOpen: true}}
			}

			if station.StationSchedule == nil {
				station.StationSchedule = &[]api.StationSchedule{schedule}
			} else {
				*station.StationSchedule = append(*station.StationSchedule, schedule)
			}
		}

		//Gorm save station, schedule, and status at this point. That's why we are taking schedule. ID only after.
		if err := tx.Save(&station).Error; err != nil {
			log.Printf("Error while saving station: %v", err)
			return err
		}

		if station.StationSchedule != nil {
			for _, schedule := range *station.StationSchedule {
				if schedule.Id != nil {
					schedulesIdsMap[*schedule.Id] = true
				}
			}
		}
	}

	var schedulesIds []int64
	for id := range schedulesIdsMap {
		schedulesIds = append(schedulesIds, id)
	}

	var otherSchedules []api.StationSchedule
	if err := tx.Not("id", schedulesIds).Where("DATE(date) >= CURRENT_DATE").Find(&otherSchedules).Error; err != nil {
		log.Printf("Error while searching other schedules: %v", err)
		return err
	}
	for _, schedule := range otherSchedules {
		if p.isScheduleToday(schedule) {
			lastStatus := api.StationStatus{}
			if err := tx.Where("station_schedule_id = ? AND user_id IS NULL", schedule.Id).Order("created_at DESC").First(&lastStatus).Error; err != nil {
				log.Printf("Error while searching status: %v", err)
				return err
			}
			if lastStatus.Id == nil || lastStatus.IsOpen {
				schedule.StationStatus = &[]api.StationStatus{{IsOpen: false}}
				if err := tx.Save(&schedule).Error; err != nil {
					log.Printf("Error while saving status: %v", err)
					return err
				}
			}
		} else if !p.isDatePassed(schedule.Date) {
			if err := tx.Delete(&schedule).Error; err != nil {
				log.Printf("Error while deleting schedule: %v", err)
				return err
			}
		}
	}
	return nil
}

func (p ScheduleDataWriter) isScheduleToday(schedule api.StationSchedule) bool {
	scheduleDate := schedule.Date.UTC().Truncate(24 * time.Hour)
	return scheduleDate.Equal(p.SinceTime)
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func (p ScheduleDataWriter) isDatePassed(date time.Time) bool {
	normalizedDate := startOfDay(date)
	normalizedSinceTime := startOfDay(p.SinceTime)

	return normalizedDate.Before(normalizedSinceTime)
}
