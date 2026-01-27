package entity

import (
	"time"

	"github.com/google/uuid"
)

// User represents the centralized authentication table
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RoleID    int       `gorm:"not null;index" json:"role_id"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"type:text;not null" json:"-"`
	FullName  string    `gorm:"type:varchar(255);not null" json:"full_name"`
	IsActive  *bool     `gorm:"not null;default:true;index" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Role           Role            `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	DoctorProfile  *DoctorProfile  `gorm:"foreignKey:UserID" json:"doctor_profile,omitempty"`
	PatientProfile *PatientProfile `gorm:"foreignKey:UserID" json:"patient_profile,omitempty"`
}

func (User) TableName() string {
	return "users"
}
