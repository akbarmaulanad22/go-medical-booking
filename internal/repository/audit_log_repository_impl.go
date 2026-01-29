package repository

import (
	"go-template-clean-architecture/internal/domain/entity"
	domainRepo "go-template-clean-architecture/internal/domain/repository"

	"gorm.io/gorm"
)

type auditLogRepository struct{}

func NewAuditLogRepository() domainRepo.AuditLogRepository {
	return &auditLogRepository{}
}

func (r *auditLogRepository) Create(db *gorm.DB, log *entity.AuditLog) error {
	return db.Create(log).Error
}

func (r *auditLogRepository) FindAll(db *gorm.DB) ([]entity.AuditLog, error) {
	var logs []entity.AuditLog
	err := db.Preload("User.Role").Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *auditLogRepository) FindByID(db *gorm.DB, id int64) (*entity.AuditLog, error) {
	var log entity.AuditLog
	err := db.Preload("User.Role").Find(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}
