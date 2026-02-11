package usecase

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
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
	ErrBookingNotFound         = errors.New("booking not found")
	ErrAlreadyBooked           = errors.New("you have already booked this schedule")
	ErrBookingAlreadyCancelled = errors.New("booking is already cancelled")
	ErrBookingNotOwned         = errors.New("booking does not belong to you")
	ErrSchedulePast            = errors.New("cannot book a past schedule")
)

type PatientBookingUsecase interface {
	GetMyBookings(ctx context.Context) (*dto.BookingListResponse, error)
	CreateBooking(ctx context.Context, req *dto.CreateBookingRequest) (*dto.BookingResponse, error)
	CancelBooking(ctx context.Context, bookingID uuid.UUID) error
}

type patientBookingUsecase struct {
	db               *gorm.DB
	log              *logrus.Logger
	bookingRepo      repository.BookingRepository
	scheduleRepo     repository.DoctorScheduleRepository
	redisSyncService *service.RedisSyncService
}

func NewPatientBookingUsecase(
	db *gorm.DB,
	log *logrus.Logger,
	bookingRepo repository.BookingRepository,
	scheduleRepo repository.DoctorScheduleRepository,
	redisSyncService *service.RedisSyncService,
) PatientBookingUsecase {
	return &patientBookingUsecase{
		db:               db,
		log:              log,
		bookingRepo:      bookingRepo,
		scheduleRepo:     scheduleRepo,
		redisSyncService: redisSyncService,
	}
}

// GetMyBookings returns all bookings for the logged-in patient
func (u *patientBookingUsecase) GetMyBookings(ctx context.Context) (*dto.BookingListResponse, error) {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		return nil, errors.New("user not found in context")
	}

	bookings, err := u.bookingRepo.FindByPatientID(u.db.WithContext(ctx), userID)
	if err != nil {
		u.log.Warnf("Failed to find bookings for patient %s: %+v", userID, err)
		return nil, err
	}

	return &dto.BookingListResponse{
		Bookings: converter.BookingsToResponses(bookings),
		Total:    len(bookings),
	}, nil
}

// CreateBooking creates a new booking with high-concurrency Redis-first approach.
//
// Flow:
// 1. Validate schedule exists and is not in the past
// 2. Check patient hasn't already booked this schedule
// 3. Redis DecrQuotaAndIncrQueue (atomic slot reservation)
// 4. Generate booking code
// 5. Insert booking to DB
// 6. If DB fails -> compensate: RestoreQuota in Redis
func (u *patientBookingUsecase) CreateBooking(ctx context.Context, req *dto.CreateBookingRequest) (*dto.BookingResponse, error) {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		return nil, errors.New("user not found in context")
	}

	// Step 1: Validate schedule exists and is active
	schedule, err := u.scheduleRepo.FindByID(u.db.WithContext(ctx), req.ScheduleID)
	if err != nil {
		u.log.Warnf("Failed to find schedule %d: %+v", req.ScheduleID, err)
		return nil, err
	}
	if schedule == nil {
		return nil, ErrScheduleNotFound
	}

	// Validate schedule is not in the past
	today := time.Now().UTC().Truncate(24 * time.Hour)
	if schedule.ScheduleDate.Before(today) {
		return nil, ErrSchedulePast
	}

	// Step 2: Check patient hasn't already booked this schedule (prevent duplicate)
	existing, err := u.bookingRepo.FindByPatientAndSchedule(u.db.WithContext(ctx), userID, req.ScheduleID)
	if err != nil {
		u.log.Warnf("Failed to check existing booking: %+v", err)
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyBooked
	}

	// Step 3: Redis atomic slot reservation (HIGH CONCURRENCY)
	// This is the critical section - thousands of users hit Redis instead of DB locks
	queueNumber, err := u.redisSyncService.DecrQuotaAndIncrQueue(ctx, req.ScheduleID)
	if err != nil {
		if errors.Is(err, service.ErrQuotaFull) {
			return nil, service.ErrQuotaFull
		}
		u.log.Warnf("Failed Redis slot reservation for schedule %d: %+v", req.ScheduleID, err)
		return nil, err
	}

	// Step 4: Generate booking code
	bookingCode := generateBookingCode(schedule.ScheduleDate)

	// Step 5: Insert booking to DB
	booking := &entity.Booking{
		PatientID:   userID,
		ScheduleID:  req.ScheduleID,
		BookingCode: bookingCode,
		QueueNumber: queueNumber,
		Status:      entity.BookingStatusPending,
	}

	if err := u.bookingRepo.Create(u.db.WithContext(ctx), booking); err != nil {
		u.log.Errorf("Failed to insert booking to DB, compensating Redis: %+v", err)

		// Step 6: COMPENSATE - restore Redis quota since DB insert failed
		syncCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if restoreErr := u.redisSyncService.RestoreQuota(syncCtx, req.ScheduleID); restoreErr != nil {
			u.log.Errorf("CRITICAL: Failed to restore Redis quota after DB failure for schedule %d: %+v", req.ScheduleID, restoreErr)
		}

		return nil, err
	}

	// Reload booking with schedule+doctor info for response
	fullBooking, err := u.bookingRepo.FindByID(u.db.WithContext(ctx), booking.ID)
	if err != nil || fullBooking == nil {
		// Return basic response if reload fails
		u.log.Warnf("Failed to reload booking %s: %+v", booking.ID, err)
		return converter.BookingToResponse(booking), nil
	}

	u.log.Infof("Booking created: id=%s, schedule=%d, queue=%d, code=%s", booking.ID, req.ScheduleID, queueNumber, bookingCode)
	return converter.BookingToResponse(fullBooking), nil
}

// CancelBooking cancels a booking and restores the schedule slot.
//
// Flow:
// 1. Find booking and verify ownership
// 2. Verify not already cancelled
// 3. Update status to cancelled in DB
// 4. RestoreQuota in Redis (does NOT decrement queue number)
func (u *patientBookingUsecase) CancelBooking(ctx context.Context, bookingID uuid.UUID) error {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		return errors.New("user not found in context")
	}

	// Step 1: Find booking
	booking, err := u.bookingRepo.FindByID(u.db.WithContext(ctx), bookingID)
	if err != nil {
		u.log.Warnf("Failed to find booking %s: %+v", bookingID, err)
		return err
	}
	if booking == nil {
		return ErrBookingNotFound
	}

	// Verify ownership
	if booking.PatientID != userID {
		return ErrBookingNotOwned
	}

	// Step 2: Check not already cancelled
	if booking.IsCancelled() {
		return ErrBookingAlreadyCancelled
	}

	// Step 3: Update status to cancelled in DB
	if err := u.bookingRepo.UpdateStatus(u.db.WithContext(ctx), bookingID, entity.BookingStatusCancelled); err != nil {
		u.log.Warnf("Failed to cancel booking %s: %+v", bookingID, err)
		return err
	}

	// Step 4: Restore quota in Redis (queue number NOT decremented)
	syncCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := u.redisSyncService.RestoreQuota(syncCtx, booking.ScheduleID); err != nil {
		// Log but don't fail - Redis will be re-synced on next startup
		u.log.Warnf("Failed to restore Redis quota for schedule %d (non-fatal): %+v", booking.ScheduleID, err)
	}

	u.log.Infof("Booking cancelled: id=%s, schedule=%d", bookingID, booking.ScheduleID)
	return nil
}

// generateBookingCode generates a unique booking code: BK-YYYYMMDD-XXXXXX
func generateBookingCode(scheduleDate time.Time) string {
	dateStr := scheduleDate.Format("20060102")
	randomBytes := make([]byte, 3)
	rand.Read(randomBytes)
	randomStr := fmt.Sprintf("%06X", randomBytes)
	return fmt.Sprintf("BK-%s-%s", dateStr, randomStr)
}
