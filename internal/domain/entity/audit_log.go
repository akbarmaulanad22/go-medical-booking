package entity

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog represents a system audit trail entry
type AuditLog struct {
	ID        int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Action    string     `gorm:"type:varchar(100);not null;index" json:"action"`
	Metadata  JSON       `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt time.Time  `gorm:"autoCreateTime;index" json:"created_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

// JSON type for GORM JSONB support
type JSON map[string]interface{}

// Common audit actions
const (
	AuditActionUserLogin      = "user.login"
	AuditActionUserLogout     = "user.logout"
	AuditActionUserRegister   = "user.register"
	AuditActionBookingCreate  = "booking.create"
	AuditActionBookingConfirm = "booking.confirm"
	AuditActionBookingCancel  = "booking.cancel"
	AuditActionScheduleCreate = "schedule.create"
	AuditActionScheduleUpdate = "schedule.update"
	AuditActionScheduleDelete = "schedule.delete"
	AuditActionProfileUpdate  = "profile.update"
)
