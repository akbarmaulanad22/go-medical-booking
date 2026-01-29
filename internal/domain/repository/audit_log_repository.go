package repository

import (
	"go-template-clean-architecture/internal/domain/entity"

	"gorm.io/gorm"
)

type AuditLogRepository interface {
	Create(db *gorm.DB, log *entity.AuditLog) error
	FindAll(db *gorm.DB) ([]entity.AuditLog, error)
	FindByID(db *gorm.DB, id int64) (*entity.AuditLog, error)
}
