package usecase

import (
	"context"
	"errors"
	"time"

	"go-template-clean-architecture/internal/delivery/dto"
	"go-template-clean-architecture/internal/domain/entity"
	"go-template-clean-architecture/internal/domain/repository"

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
	db                *gorm.DB
	log               *logrus.Logger
	scheduleRepo      repository.DoctorScheduleRepository
	doctorProfileRepo repository.DoctorProfileRepository
}

func NewDoctorScheduleUsecase(
	db *gorm.DB,
	log *logrus.Logger,
	scheduleRepo repository.DoctorScheduleRepository,
	doctorProfileRepo repository.DoctorProfileRepository,
) DoctorScheduleUsecase {
	return &doctorScheduleUsecase{
		db:                db,
		log:               log,
		scheduleRepo:      scheduleRepo,
		doctorProfileRepo: doctorProfileRepo,
	}
}

func (u *doctorScheduleUsecase) CreateSchedule(ctx context.Context, req *dto.CreateScheduleRequest) (*dto.ScheduleResponse, error) {
	// Validate doctor exists
	doctor, err := u.doctorProfileRepo.FindByUserID(u.db, req.DoctorID)
	if err != nil {
		u.log.Warnf("Failed to find doctor: %+v", err)
		return nil, err
	}
	if doctor == nil {
		return nil, ErrDoctorNotFound
	}

	// Parse schedule date
	scheduleDate, err := time.Parse("2006-01-02", req.ScheduleDate)
	if err != nil {
		return nil, ErrInvalidScheduleDate
	}

	// Validate time format
	if _, err := time.Parse("15:04", req.StartTime); err != nil {
		return nil, ErrInvalidTimeFormat
	}
	if _, err := time.Parse("15:04", req.EndTime); err != nil {
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

	if err := u.scheduleRepo.Create(ctx, u.db, schedule); err != nil {
		u.log.Warnf("Failed to create schedule: %+v", err)
		return nil, err
	}

	return &dto.ScheduleResponse{
		ID:             schedule.ID,
		DoctorID:       schedule.DoctorID,
		ScheduleDate:   schedule.ScheduleDate.Format("2006-01-02"),
		StartTime:      schedule.StartTime,
		EndTime:        schedule.EndTime,
		TotalQuota:     schedule.TotalQuota,
		RemainingQuota: schedule.RemainingQuota,
		CreatedAt:      schedule.CreatedAt,
		UpdatedAt:      schedule.UpdatedAt,
	}, nil
}

func (u *doctorScheduleUsecase) GetSchedule(ctx context.Context, scheduleID int) (*dto.ScheduleResponse, error) {
	schedule, err := u.scheduleRepo.FindByID(ctx, u.db, scheduleID)
	if err != nil {
		u.log.Warnf("Failed to find schedule: %+v", err)
		return nil, err
	}
	if schedule == nil {
		return nil, ErrScheduleNotFound
	}

	response := &dto.ScheduleResponse{
		ID:             schedule.ID,
		DoctorID:       schedule.DoctorID,
		ScheduleDate:   schedule.ScheduleDate.Format("2006-01-02"),
		StartTime:      schedule.StartTime,
		EndTime:        schedule.EndTime,
		TotalQuota:     schedule.TotalQuota,
		RemainingQuota: schedule.RemainingQuota,
		CreatedAt:      schedule.CreatedAt,
		UpdatedAt:      schedule.UpdatedAt,
	}

	// Include doctor info if available
	if schedule.Doctor.UserID != uuid.Nil {
		response.Doctor = &dto.DoctorResponse{
			ID:             schedule.Doctor.UserID,
			Email:          schedule.Doctor.User.Email,
			FullName:       schedule.Doctor.User.FullName,
			STRNumber:      schedule.Doctor.STRNumber,
			Specialization: schedule.Doctor.Specialization,
			Biography:      schedule.Doctor.Biography,
			IsActive:       schedule.Doctor.User.IsActive,
		}
	}

	return response, nil
}

func (u *doctorScheduleUsecase) GetSchedulesByDoctor(ctx context.Context, doctorID uuid.UUID) (*dto.ScheduleListResponse, error) {
	schedules, err := u.scheduleRepo.FindByDoctorID(ctx, u.db, doctorID)
	if err != nil {
		u.log.Warnf("Failed to find schedules: %+v", err)
		return nil, err
	}

	responses := make([]dto.ScheduleResponse, len(schedules))
	for i, schedule := range schedules {
		responses[i] = dto.ScheduleResponse{
			ID:             schedule.ID,
			DoctorID:       schedule.DoctorID,
			ScheduleDate:   schedule.ScheduleDate.Format("2006-01-02"),
			StartTime:      schedule.StartTime,
			EndTime:        schedule.EndTime,
			TotalQuota:     schedule.TotalQuota,
			RemainingQuota: schedule.RemainingQuota,
			CreatedAt:      schedule.CreatedAt,
			UpdatedAt:      schedule.UpdatedAt,
		}
	}

	return &dto.ScheduleListResponse{
		Schedules: responses,
		Total:     len(responses),
	}, nil
}

func (u *doctorScheduleUsecase) GetAllSchedules(ctx context.Context) (*dto.ScheduleListResponse, error) {
	schedules, err := u.scheduleRepo.FindAll(ctx, u.db)
	if err != nil {
		u.log.Warnf("Failed to find all schedules: %+v", err)
		return nil, err
	}

	responses := make([]dto.ScheduleResponse, len(schedules))
	for i, schedule := range schedules {
		response := dto.ScheduleResponse{
			ID:             schedule.ID,
			DoctorID:       schedule.DoctorID,
			ScheduleDate:   schedule.ScheduleDate.Format("2006-01-02"),
			StartTime:      schedule.StartTime,
			EndTime:        schedule.EndTime,
			TotalQuota:     schedule.TotalQuota,
			RemainingQuota: schedule.RemainingQuota,
			CreatedAt:      schedule.CreatedAt,
			UpdatedAt:      schedule.UpdatedAt,
		}

		if schedule.Doctor.UserID != uuid.Nil {
			response.Doctor = &dto.DoctorResponse{
				ID:             schedule.Doctor.UserID,
				Email:          schedule.Doctor.User.Email,
				FullName:       schedule.Doctor.User.FullName,
				STRNumber:      schedule.Doctor.STRNumber,
				Specialization: schedule.Doctor.Specialization,
				Biography:      schedule.Doctor.Biography,
				IsActive:       schedule.Doctor.User.IsActive,
			}
		}

		responses[i] = response
	}

	return &dto.ScheduleListResponse{
		Schedules: responses,
		Total:     len(responses),
	}, nil
}

func (u *doctorScheduleUsecase) UpdateSchedule(ctx context.Context, scheduleID int, req *dto.UpdateScheduleRequest) (*dto.ScheduleResponse, error) {
	schedule, err := u.scheduleRepo.FindByID(ctx, u.db, scheduleID)
	if err != nil {
		u.log.Warnf("Failed to find schedule: %+v", err)
		return nil, err
	}
	if schedule == nil {
		return nil, ErrScheduleNotFound
	}

	// Update fields
	if req.ScheduleDate != "" {
		scheduleDate, err := time.Parse("2006-01-02", req.ScheduleDate)
		if err != nil {
			return nil, ErrInvalidScheduleDate
		}
		schedule.ScheduleDate = scheduleDate
	}
	if req.StartTime != "" {
		if _, err := time.Parse("15:04", req.StartTime); err != nil {
			return nil, ErrInvalidTimeFormat
		}
		schedule.StartTime = req.StartTime
	}
	if req.EndTime != "" {
		if _, err := time.Parse("15:04", req.EndTime); err != nil {
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

	if err := u.scheduleRepo.Update(ctx, u.db, schedule); err != nil {
		u.log.Warnf("Failed to update schedule: %+v", err)
		return nil, err
	}

	return &dto.ScheduleResponse{
		ID:             schedule.ID,
		DoctorID:       schedule.DoctorID,
		ScheduleDate:   schedule.ScheduleDate.Format("2006-01-02"),
		StartTime:      schedule.StartTime,
		EndTime:        schedule.EndTime,
		TotalQuota:     schedule.TotalQuota,
		RemainingQuota: schedule.RemainingQuota,
		CreatedAt:      schedule.CreatedAt,
		UpdatedAt:      schedule.UpdatedAt,
	}, nil
}

func (u *doctorScheduleUsecase) DeleteSchedule(ctx context.Context, scheduleID int) error {
	schedule, err := u.scheduleRepo.FindByID(ctx, u.db, scheduleID)
	if err != nil {
		u.log.Warnf("Failed to find schedule: %+v", err)
		return err
	}
	if schedule == nil {
		return ErrScheduleNotFound
	}

	if err := u.scheduleRepo.Delete(ctx, u.db, scheduleID); err != nil {
		u.log.Warnf("Failed to delete schedule: %+v", err)
		return err
	}

	return nil
}
