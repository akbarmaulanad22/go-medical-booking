package entity

// Role represents a user role in the system
type Role struct {
	ID          int    `gorm:"primaryKey;autoIncrement" json:"id"`
	RoleName    string `gorm:"type:varchar(50);uniqueIndex;not null" json:"role_name"`
	Description string `gorm:"type:text" json:"description,omitempty"`

	// Relationships
	Users []User `gorm:"foreignKey:RoleID" json:"users,omitempty"`
}

func (Role) TableName() string {
	return "roles"
}

// Role ID constants
const (
	RoleIDAdmin   = 1
	RoleIDDoctor  = 2
	RoleIDPatient = 3
)

// RoleNames constants
const (
	RoleAdmin   = "admin"
	RoleDoctor  = "doctor"
	RolePatient = "patient"
)
