package service

import (
	"context"

	"go-template-clean-architecture/internal/domain/entity"
	"go-template-clean-architecture/internal/domain/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AuditService interface {
	LogCreate(ctx context.Context, tx *gorm.DB, userID *uuid.UUID, action string, entityName string, entityID string, newValue interface{}) error
	LogUpdate(ctx context.Context, tx *gorm.DB, userID *uuid.UUID, action string, entityName string, entityID string, oldValue, newValue interface{}) error
	LogDelete(ctx context.Context, tx *gorm.DB, userID *uuid.UUID, action string, entityName string, entityID string, oldValue interface{}) error
}

type auditService struct {
	db        *gorm.DB
	log       *logrus.Logger
	auditRepo repository.AuditLogRepository
}

func NewAuditService(db *gorm.DB, log *logrus.Logger, auditRepo repository.AuditLogRepository) AuditService {
	return &auditService{
		db:        db,
		log:       log,
		auditRepo: auditRepo,
	}
}

// LogCreate logs a create action
func (s *auditService) LogCreate(ctx context.Context, tx *gorm.DB, userID *uuid.UUID, action string, entityName string, entityID string, newValue interface{}) error {
	metadata := entity.JSON{
		"entity":    entityName,
		"entity_id": entityID,
		"old_value": nil,
		"new_value": newValue,
	}

	auditLog := &entity.AuditLog{
		UserID:   userID,
		Action:   action,
		Metadata: metadata,
	}

	if err := s.auditRepo.Create(tx, auditLog); err != nil {
		s.log.Warnf("Failed to create audit log: %+v", err)
		return err
	}

	return nil
}

// LogUpdate logs an update action with old and new values
func (s *auditService) LogUpdate(ctx context.Context, tx *gorm.DB, userID *uuid.UUID, action string, entityName string, entityID string, oldValue, newValue interface{}) error {
	metadata := entity.JSON{
		"entity":    entityName,
		"entity_id": entityID,
		"old_value": oldValue,
		"new_value": newValue,
	}

	auditLog := &entity.AuditLog{
		UserID:   userID,
		Action:   action,
		Metadata: metadata,
	}

	if err := s.auditRepo.Create(tx, auditLog); err != nil {
		s.log.Warnf("Failed to create audit log: %+v", err)
		return err
	}

	return nil
}

// LogDelete logs a delete action with old value
func (s *auditService) LogDelete(ctx context.Context, tx *gorm.DB, userID *uuid.UUID, action string, entityName string, entityID string, oldValue interface{}) error {
	metadata := entity.JSON{
		"entity":    entityName,
		"entity_id": entityID,
		"old_value": oldValue,
		"new_value": nil,
	}

	auditLog := &entity.AuditLog{
		UserID:   userID,
		Action:   action,
		Metadata: metadata,
	}

	if err := s.auditRepo.Create(tx, auditLog); err != nil {
		s.log.Warnf("Failed to create audit log: %+v", err)
		return err
	}

	return nil
}
