package entity

import "github.com/google/uuid"

// DoctorProfile represents doctor-specific profile data
type DoctorProfile struct {
	UserID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	STRNumber      string    `gorm:"column:str_number;type:varchar(50);uniqueIndex;not null" json:"str_number"`
	Specialization string    `gorm:"type:varchar(100);not null;index" json:"specialization"`
	Biography      string    `gorm:"type:text" json:"biography,omitempty"`

	// Relationships
	User      User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Schedules []DoctorSchedule `gorm:"foreignKey:DoctorID" json:"schedules,omitempty"`
}

func (DoctorProfile) TableName() string {
	return "doctor_profiles"
}
