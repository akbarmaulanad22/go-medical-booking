package entity

import (
	"time"

	"github.com/google/uuid"
)

// DoctorSchedule represents doctor availability with quota management
// Note: RemainingQuota is calculated from Redis/DB query, not stored in entity
type DoctorSchedule struct {
	ID           int       `gorm:"primaryKey;autoIncrement" json:"id"`
	DoctorID     uuid.UUID `gorm:"type:uuid;not null;index" json:"doctor_id"`
	ScheduleDate time.Time `gorm:"type:date;not null;index" json:"schedule_date"`
	StartTime    string    `gorm:"type:time;not null" json:"start_time"`
	EndTime      string    `gorm:"type:time;not null" json:"end_time"`
	TotalQuota   int       `gorm:"not null" json:"total_quota"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Doctor   DoctorProfile `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
	Bookings []Booking     `gorm:"foreignKey:ScheduleID" json:"bookings,omitempty"`
}

func (DoctorSchedule) TableName() string {
	return "doctor_schedules"
}
