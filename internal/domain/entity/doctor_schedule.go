package entity

import (
	"time"

	"github.com/google/uuid"
)

// DoctorSchedule represents doctor availability with quota management
type DoctorSchedule struct {
	ID             int       `gorm:"primaryKey;autoIncrement" json:"id"`
	DoctorID       uuid.UUID `gorm:"type:uuid;not null;index" json:"doctor_id"`
	ScheduleDate   time.Time `gorm:"type:date;not null;index" json:"schedule_date"`
	StartTime      string    `gorm:"type:time;not null" json:"start_time"`
	EndTime        string    `gorm:"type:time;not null" json:"end_time"`
	TotalQuota     int       `gorm:"not null" json:"total_quota"`
	RemainingQuota int       `gorm:"not null" json:"remaining_quota"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Doctor   DoctorProfile `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
	Bookings []Booking     `gorm:"foreignKey:ScheduleID" json:"bookings,omitempty"`
}

func (DoctorSchedule) TableName() string {
	return "doctor_schedules"
}

// HasAvailableSlot checks if the schedule has remaining quota
func (ds *DoctorSchedule) HasAvailableSlot() bool {
	return ds.RemainingQuota > 0
}

// DecrementQuota decreases the remaining quota by 1
// Note: For high-concurrency, use database-level atomic operations with SELECT FOR UPDATE
func (ds *DoctorSchedule) DecrementQuota() bool {
	if ds.RemainingQuota > 0 {
		ds.RemainingQuota--
		return true
	}
	return false
}

// IncrementQuota increases the remaining quota by 1 (for cancellation)
func (ds *DoctorSchedule) IncrementQuota() bool {
	if ds.RemainingQuota < ds.TotalQuota {
		ds.RemainingQuota++
		return true
	}
	return false
}
