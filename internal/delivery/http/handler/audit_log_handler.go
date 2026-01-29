package handler

import (
	"net/http"
	"strconv"

	"go-template-clean-architecture/internal/usecase"
	"go-template-clean-architecture/pkg/response"

	"github.com/gorilla/mux"
)

type AuditLogHandler struct {
	auditLogUsecase usecase.AuditLogUsecase
}

func NewAuditLogHandler(auditLogUsecase usecase.AuditLogUsecase) *AuditLogHandler {
	return &AuditLogHandler{
		auditLogUsecase: auditLogUsecase,
	}
}

func (h *AuditLogHandler) GetAuditLog(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auditLogID, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid audit log ID", nil)
		return
	}

	auditLog, err := h.auditLogUsecase.GetAuditLog(r.Context(), int64(auditLogID))
	if err != nil {
		if err == usecase.ErrAuditLogNotFound {
			response.NotFound(w, "Audit log not found")
			return
		}
		response.InternalServerError(w, "Failed to get audit log")
		return
	}

	response.Success(w, http.StatusOK, "Audit log retrieved successfully", auditLog)
}

func (h *AuditLogHandler) GetAllAuditLogs(w http.ResponseWriter, r *http.Request) {
	auditLogs, err := h.auditLogUsecase.GetAllAuditLogs(r.Context())
	if err != nil {
		response.InternalServerError(w, "Failed to get audit logs")
		return
	}

	response.Success(w, http.StatusOK, "Audit logs retrieved successfully", auditLogs)
}
