package usecase

import (
	"context"
	"errors"
	"strconv"
	"time"

	"go-template-clean-architecture/internal/converter"
	"go-template-clean-architecture/internal/delivery/dto"
	"go-template-clean-architecture/internal/delivery/http/middleware"
	"go-template-clean-architecture/internal/domain/entity"
	"go-template-clean-architecture/internal/domain/repository"
	"go-template-clean-architecture/internal/service"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrScheduleNotFound    = errors.New("schedule not found")
	ErrInvalidScheduleDate = errors.New("invalid schedule date format, use YYYY-MM-DD")
	ErrInvalidTimeFormat   = errors.New("invalid time format, use HH:MM")
)

type DoctorScheduleUsecase interface {
	CreateSchedule(ctx context.Context, req *dto.CreateScheduleRequest) (*dto.ScheduleResponse, error)
	GetSchedule(ctx context.Context, scheduleID int) (*dto.ScheduleResponse, error)
	GetSchedulesByDoctor(ctx context.Context, doctorID uuid.UUID) (*dto.ScheduleListResponse, error)
	GetAllSchedules(ctx context.Context) (*dto.ScheduleListResponse, error)
	UpdateSchedule(ctx context.Context, scheduleID int, req *dto.UpdateScheduleRequest) (*dto.ScheduleResponse, error)
	DeleteSchedule(ctx context.Context, scheduleID int) error
}

type doctorScheduleUsecase struct {
	db               *gorm.DB
	log              *logrus.Logger
	scheduleRepo     repository.DoctorScheduleRepository
	auditService     service.AuditService
	redisSyncService *service.RedisSyncService
}

func NewDoctorScheduleUsecase(
	db *gorm.DB,
	log *logrus.Logger,
	scheduleRepo repository.DoctorScheduleRepository,
	auditService service.AuditService,
	redisSyncService *service.RedisSyncService,
) DoctorScheduleUsecase {
	return &doctorScheduleUsecase{
		db:               db,
		log:              log,
		scheduleRepo:     scheduleRepo,
		auditService:     auditService,
		redisSyncService: redisSyncService,
	}
}

// CreateSchedule creates a new doctor schedule and syncs to Redis SYNCHRONOUSLY.
//
// Sync Strategy:
// - After DB commit, calls SyncScheduleQuota synchronously (no goroutine)
// - Redis sync failure is logged but does not rollback DB (fail-safe)
// - Admin reliability > speed, so we wait for Redis response
func (u *doctorScheduleUsecase) CreateSchedule(ctx context.Context, req *dto.CreateScheduleRequest) (*dto.ScheduleResponse, error) {
	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Parse schedule date
	scheduleDate, err := time.Parse("2006-01-02", req.ScheduleDate)
	if err != nil {
		u.log.Warnf("Failed to parse schedule date: %+v", err)
		return nil, ErrInvalidScheduleDate
	}

	// Validate time format
	if _, err := time.Parse("15:04", req.StartTime); err != nil {
		u.log.Warnf("Failed to parse start time: %+v", err)
		return nil, ErrInvalidTimeFormat
	}
	if _, err := time.Parse("15:04", req.EndTime); err != nil {
		u.log.Warnf("Failed to parse end time: %+v", err)
		return nil, ErrInvalidTimeFormat
	}

	schedule := &entity.DoctorSchedule{
		DoctorID:     req.DoctorID,
		ScheduleDate: scheduleDate,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		TotalQuota:   req.TotalQuota,
	}

	if err := u.scheduleRepo.Create(tx, schedule); err != nil {
		u.log.Warnf("Failed to create schedule: %+v", err)
		if isForeignKeyError(err, "doctor") {
			return nil, ErrDoctorNotFound
		}
		return nil, err
	}

	// Audit log - create schedule
	userID, _ := middleware.GetUserIDFromContext(ctx)
	if err := u.auditService.LogCreate(ctx, tx, &userID, entity.AuditActionScheduleCreate, "doctor_schedule", strconv.Itoa(schedule.ID), converter.ScheduleToResponse(schedule)); err != nil {
		u.log.Warnf("Failed to create audit log: %+v", err)
	}

	if err := tx.Commit().Error; err != nil {
		u.log.Warnf("Failed commit transaction: %+v", err)
		return nil, err
	}

	// SYNCHRONOUS Redis sync - no goroutine
	// Reliability > Speed for Admin operations
	syncCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := u.redisSyncService.SyncScheduleQuota(syncCtx, schedule.ID, schedule.TotalQuota, schedule.ScheduleDate); err != nil {
		// Log error but don't fail the request (fail-safe)
		// Redis will be synced on next startup or manual trigger
		u.log.Warnf("Redis sync failed for new schedule %d (non-fatal): %+v", schedule.ID, err)
	} else {
		u.log.Infof("Schedule %d created and synced to Redis", schedule.ID)
	}

	return converter.ScheduleToResponse(schedule), nil
}

func (u *doctorScheduleUsecase) GetSchedule(ctx context.Context, scheduleID int) (*dto.ScheduleResponse, error) {
	schedule, err := u.scheduleRepo.FindByID(u.db, scheduleID)
	if err != nil {
		u.log.Warnf("Failed to find schedule: %+v", err)
		return nil, err
	}
	if schedule == nil {
		u.log.Warnf("Schedule not found")
		return nil, ErrScheduleNotFound
	}

	return converter.ScheduleToResponse(schedule), nil
}

func (u *doctorScheduleUsecase) GetSchedulesByDoctor(ctx context.Context, doctorID uuid.UUID) (*dto.ScheduleListResponse, error) {
	schedules, err := u.scheduleRepo.FindByDoctorID(u.db, doctorID)
	if err != nil {
		u.log.Warnf("Failed to find schedules: %+v", err)
		return nil, err
	}

	return &dto.ScheduleListResponse{
		Schedules: converter.SchedulesToResponses(schedules),
		Total:     len(schedules),
	}, nil
}

func (u *doctorScheduleUsecase) GetAllSchedules(ctx context.Context) (*dto.ScheduleListResponse, error) {
	schedules, err := u.scheduleRepo.FindAll(u.db)
	if err != nil {
		u.log.Warnf("Failed to find all schedules: %+v", err)
		return nil, err
	}

	return &dto.ScheduleListResponse{
		Schedules: converter.SchedulesToResponses(schedules),
		Total:     len(schedules),
	}, nil
}

// UpdateSchedule updates a schedule and syncs to Redis SYNCHRONOUSLY.
//
// Delta Strategy for TotalQuota changes:
// - If TotalQuota changes, calculate delta = NewTotal - OldTotal
// - Use Redis INCRBY(delta) instead of SET(absoluteValue)
// - This prevents race condition if user books at exact same millisecond
//
// Sync Strategy:
// - Synchronous (no goroutine) - reliability > speed for Admin
func (u *doctorScheduleUsecase) UpdateSchedule(ctx context.Context, scheduleID int, req *dto.UpdateScheduleRequest) (*dto.ScheduleResponse, error) {
	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	schedule, err := u.scheduleRepo.FindByID(tx, scheduleID)
	if err != nil {
		u.log.Warnf("Failed to find schedule: %+v", err)
		return nil, err
	}
	if schedule == nil {
		u.log.Warnf("Schedule not found")
		return nil, ErrScheduleNotFound
	}

	// Capture old values for audit and delta calculation
	oldValue := converter.ScheduleToResponse(schedule)
	oldTotalQuota := schedule.TotalQuota
	oldScheduleDate := schedule.ScheduleDate

	// Update fields
	if req.DoctorID != uuid.Nil {
		schedule.DoctorID = req.DoctorID
	}

	if req.ScheduleDate != "" {
		scheduleDate, err := time.Parse("2006-01-02", req.ScheduleDate)
		if err != nil {
			u.log.Warnf("Failed to parse schedule date: %+v", err)
			return nil, ErrInvalidScheduleDate
		}
		schedule.ScheduleDate = scheduleDate
	}
	if req.StartTime != "" {
		if _, err := time.Parse("15:04", req.StartTime); err != nil {
			u.log.Warnf("Failed to parse start time: %+v", err)
			return nil, ErrInvalidTimeFormat
		}
		schedule.StartTime = req.StartTime
	}
	if req.EndTime != "" {
		if _, err := time.Parse("15:04", req.EndTime); err != nil {
			u.log.Warnf("Failed to parse end time: %+v", err)
			return nil, ErrInvalidTimeFormat
		}
		schedule.EndTime = req.EndTime
	}

	// Handle TotalQuota change with delta strategy
	var quotaDelta int
	quotaChanged := false

	if req.TotalQuota != nil && *req.TotalQuota != oldTotalQuota {
		quotaDelta = *req.TotalQuota - oldTotalQuota
		quotaChanged = true

		schedule.TotalQuota = *req.TotalQuota
	}

	if err := u.scheduleRepo.Update(tx, schedule); err != nil {
		u.log.Warnf("Failed to update schedule: %+v", err)
		if isForeignKeyError(err, "doctor") {
			return nil, ErrDoctorNotFound
		}
		return nil, err
	}

	// Audit log - update schedule
	newValue := converter.ScheduleToResponse(schedule)
	userID, _ := middleware.GetUserIDFromContext(ctx)
	if err := u.auditService.LogUpdate(ctx, tx, &userID, entity.AuditActionScheduleUpdate, "doctor_schedule", strconv.Itoa(scheduleID), oldValue, newValue); err != nil {
		u.log.Warnf("Failed to create audit log: %+v", err)
	}

	if err := tx.Commit().Error; err != nil {
		u.log.Warnf("Failed commit transaction: %+v", err)
		return nil, err
	}

	// SYNCHRONOUS Redis sync - no goroutine
	// Use detached context so Redis sync is not cancelled by HTTP request timeout
	syncCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Handle different update scenarios
	dateChanged := !schedule.ScheduleDate.Equal(oldScheduleDate)

	if dateChanged {
		// Schedule date changed - delete old keys and create new ones
		u.log.Infof("Schedule %d date changed, re-syncing Redis keys", scheduleID)

		// Delete old keys (they had TTL based on old date)
		if err := u.redisSyncService.DeleteScheduleKeys(syncCtx, scheduleID); err != nil {
			u.log.Warnf("Failed to delete old Redis keys for schedule %d (non-fatal): %+v", scheduleID, err)
		}

		// Create new keys with new TTL
		if err := u.redisSyncService.SyncScheduleQuota(syncCtx, scheduleID, schedule.TotalQuota, schedule.ScheduleDate); err != nil {
			u.log.Warnf("Failed to sync new Redis keys for schedule %d (non-fatal): %+v", scheduleID, err)
		}
	} else if quotaChanged {
		// Only quota changed - use INCRBY delta strategy
		// This prevents race condition with concurrent bookings
		if err := u.redisSyncService.UpdateScheduleQuotaDelta(syncCtx, scheduleID, quotaDelta, schedule.ScheduleDate); err != nil {
			u.log.Warnf("Failed to update Redis quota for schedule %d (non-fatal): %+v", scheduleID, err)
		} else {
			u.log.Infof("Schedule %d quota updated by delta %d", scheduleID, quotaDelta)
		}
	}

	return converter.ScheduleToResponse(schedule), nil
}

// DeleteSchedule deletes a schedule and removes Redis keys SYNCHRONOUSLY.
//
// Sync Strategy:
// - After DB commit, calls DeleteScheduleKeys synchronously
// - Redis cleanup failure is logged but does not fail request (fail-safe)
func (u *doctorScheduleUsecase) DeleteSchedule(ctx context.Context, scheduleID int) error {
	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Fetch schedule for audit log
	schedule, err := u.scheduleRepo.FindByID(tx, scheduleID)
	if err != nil {
		u.log.Warnf("Failed to find schedule for delete: %+v", err)
		return err
	}

	var oldValue *dto.ScheduleResponse
	if schedule != nil {
		oldValue = converter.ScheduleToResponse(schedule)
	}

	deleted, err := u.scheduleRepo.Delete(tx, scheduleID)
	if err != nil {
		u.log.Warnf("Failed to delete schedule: %+v", err)
		return err
	}

	if deleted == 0 {
		u.log.Warnf("Schedule not found")
		return ErrScheduleNotFound
	}

	// Audit log - delete schedule
	if oldValue != nil {
		userID, _ := middleware.GetUserIDFromContext(ctx)
		if err := u.auditService.LogDelete(ctx, tx, &userID, entity.AuditActionScheduleDelete, "doctor_schedule", strconv.Itoa(scheduleID), oldValue); err != nil {
			u.log.Warnf("Failed to create audit log: %+v", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		u.log.Warnf("Failed commit transaction: %+v", err)
		return err
	}

	// SYNCHRONOUS Redis cleanup - no goroutine
	// Use detached context so Redis cleanup is not cancelled by HTTP request timeout
	syncCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := u.redisSyncService.DeleteScheduleKeys(syncCtx, scheduleID); err != nil {
		// Log error but don't fail (fail-safe)
		// Keys will expire via TTL anyway
		u.log.Warnf("Failed to delete Redis keys for schedule %d (non-fatal): %+v", scheduleID, err)
	} else {
		u.log.Infof("Schedule %d deleted and Redis keys removed", scheduleID)
	}

	return nil
}
