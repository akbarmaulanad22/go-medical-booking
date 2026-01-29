package dto

import (
	"go-template-clean-architecture/internal/domain/entity"
	"time"
)

// Response DTOs

type AuditLogResponse struct {
	ID        int64        `json:"id"`
	User      UserResponse `json:"user"`
	Action    string       `json:"action"`
	Metadata  entity.JSON  `json:"metadata"`
	CreatedAt time.Time    `json:"created_at"`
}

type AuditLogListResponse struct {
	Logs  []AuditLogResponse `json:"logs"`
	Total int                `json:"total"`
}
