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
	db           *gorm.DB
	log          *logrus.Logger
	scheduleRepo repository.DoctorScheduleRepository
	auditService service.AuditService
}

func NewDoctorScheduleUsecase(
	db *gorm.DB,
	log *logrus.Logger,
	scheduleRepo repository.DoctorScheduleRepository,
	auditService service.AuditService,
) DoctorScheduleUsecase {
	return &doctorScheduleUsecase{
		db:           db,
		log:          log,
		scheduleRepo: scheduleRepo,
		auditService: auditService,
	}
}

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
		DoctorID:       req.DoctorID,
		ScheduleDate:   scheduleDate,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		TotalQuota:     req.TotalQuota,
		RemainingQuota: req.TotalQuota,
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

	// Capture old value for audit
	oldValue := converter.ScheduleToResponse(schedule)

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
	if req.TotalQuota != nil {
		// Adjust remaining quota proportionally
		diff := *req.TotalQuota - schedule.TotalQuota
		schedule.TotalQuota = *req.TotalQuota
		schedule.RemainingQuota += diff
		if schedule.RemainingQuota < 0 {
			schedule.RemainingQuota = 0
		}
		if schedule.RemainingQuota > schedule.TotalQuota {
			schedule.RemainingQuota = schedule.TotalQuota
		}
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

	return converter.ScheduleToResponse(schedule), nil
}

func (u *doctorScheduleUsecase) DeleteSchedule(ctx context.Context, scheduleID int) error {
	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Capture old value for audit
	// We need to fetch it first if we weren't doing a delete, but DeleteSchedule doesn't fetch first in the original code...
	// Original code:
	// deleted, err := u.scheduleRepo.Delete(tx, scheduleID)

	// Wait, if I want to log the deleted item, I MUST fetch it first.
	// The original code does `u.scheduleRepo.Delete` which returns rows affected.
	// It does NOT fetch.
	// So I need to fetch it first.

	schedule, err := u.scheduleRepo.FindByID(tx, scheduleID)
	if err != nil {
		u.log.Warnf("Failed to find schedule for delete: %+v", err)
		// Continue to delete attempt or return error?
		// If DB error, return error.
		return err
	}
	// If nil, Delete will return 0 rows anyway, but better to handle it.
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

	return nil
}
