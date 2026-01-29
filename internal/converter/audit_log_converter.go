package converter

import (
	"go-template-clean-architecture/internal/delivery/dto"
	"go-template-clean-architecture/internal/domain/entity"
)

// AuditLogToResponse converts a AuditLog entity to AuditLogResponse DTO
func AuditLogToResponse(log *entity.AuditLog) *dto.AuditLogResponse {
	if log == nil {
		return nil
	}

	return &dto.AuditLogResponse{
		ID:        log.ID,
		User:      *UserToResponse(log.User),
		Action:    log.Action,
		Metadata:  log.Metadata,
		CreatedAt: log.CreatedAt,
	}
}

// AuditLogsToResponses converts a slice of AuditLog entities to slice of AuditLogResponse DTOs
func AuditLogsToResponses(logs []entity.AuditLog) []dto.AuditLogResponse {
	responses := make([]dto.AuditLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = dto.AuditLogResponse{
			ID:        log.ID,
			User:      *UserToResponse(log.User),
			Action:    log.Action,
			Metadata:  log.Metadata,
			CreatedAt: log.CreatedAt,
		}
	}
	return responses
}
