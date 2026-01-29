package usecase

import (
	"context"
	"errors"

	"go-template-clean-architecture/internal/converter"
	"go-template-clean-architecture/internal/delivery/dto"
	"go-template-clean-architecture/internal/domain/repository"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrAuditLogNotFound = errors.New("audit log not found")
)

type AuditLogUsecase interface {
	GetAllAuditLogs(ctx context.Context) (*dto.AuditLogListResponse, error)
	GetAuditLog(ctx context.Context, id int64) (*dto.AuditLogResponse, error)
}

type auditLogUsecase struct {
	db           *gorm.DB
	log          *logrus.Logger
	auditLogRepo repository.AuditLogRepository
}

func NewAuditLogUsecase(
	db *gorm.DB,
	log *logrus.Logger,
	auditLogRepo repository.AuditLogRepository,
) AuditLogUsecase {
	return &auditLogUsecase{
		db:           db,
		log:          log,
		auditLogRepo: auditLogRepo,
	}
}

func (u *auditLogUsecase) GetAllAuditLogs(ctx context.Context) (*dto.AuditLogListResponse, error) {
	logs, err := u.auditLogRepo.FindAll(u.db)
	if err != nil {
		u.log.Warnf("Failed to find all audit logs: %+v", err)
		return nil, err
	}

	logResponses := converter.AuditLogsToResponses(logs)

	return &dto.AuditLogListResponse{
		Logs:  logResponses,
		Total: len(logs),
	}, nil
}

func (u *auditLogUsecase) GetAuditLog(ctx context.Context, id int64) (*dto.AuditLogResponse, error) {
	auditLog, err := u.auditLogRepo.FindByID(u.db, id)
	if err != nil {
		u.log.Warnf("Failed to find log audit log: %+v", err)
		return nil, err
	}
	if auditLog == nil {
		u.log.Warnf("Failed to find log audit log: %+v", "audit log not found")
		return nil, ErrDoctorNotFound
	}

	return converter.AuditLogToResponse(auditLog), nil
}
