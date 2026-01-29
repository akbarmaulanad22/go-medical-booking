package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
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

// Value returns json value, implement driver.Valuer interface
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := map[string]interface{}{}
	err := json.Unmarshal(bytes, &result)
	*j = JSON(result)
	return err
}

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
	AuditActionDoctorCreate   = "doctor.create"
	AuditActionDoctorUpdate   = "doctor.update"
	AuditActionDoctorDelete   = "doctor.delete"
)
