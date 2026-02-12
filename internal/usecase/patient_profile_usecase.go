package usecase

import (
	"context"
	"errors"

	"go-template-clean-architecture/internal/converter"
	"go-template-clean-architecture/internal/delivery/dto"
	"go-template-clean-architecture/internal/delivery/http/middleware"
	"go-template-clean-architecture/internal/domain/entity"
	"go-template-clean-architecture/internal/domain/repository"
	"go-template-clean-architecture/internal/service"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrPatientNotFound = errors.New("patient profile not found")
)

type PatientProfileUsecase interface {
	UpdateSelfProfile(ctx context.Context, req *dto.PatientUpdateSelfRequest) (*dto.PatientResponse, error)
}

type patientProfileUsecase struct {
	db                 *gorm.DB
	log                *logrus.Logger
	userRepo           repository.UserRepository
	patientProfileRepo repository.PatientProfileRepository
	auditService       service.AuditService
}

func NewPatientProfileUsecase(
	db *gorm.DB,
	log *logrus.Logger,
	userRepo repository.UserRepository,
	patientProfileRepo repository.PatientProfileRepository,
	auditService service.AuditService,
) PatientProfileUsecase {
	return &patientProfileUsecase{
		db:                 db,
		log:                log,
		userRepo:           userRepo,
		patientProfileRepo: patientProfileRepo,
		auditService:       auditService,
	}
}

// UpdateSelfProfile updates the patient's own profile.
//
// Allowed fields: password (with old password verification), phone_number, address.
// Sensitive fields (NIK, gender, date_of_birth) are NOT editable by the patient.
func (u *patientProfileUsecase) UpdateSelfProfile(ctx context.Context, req *dto.PatientUpdateSelfRequest) (*dto.PatientResponse, error) {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		return nil, errors.New("user not found in context")
	}

	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Get patient profile with user data
	profile, err := u.patientProfileRepo.FindByUserID(ctx, tx, userID)
	if err != nil {
		u.log.Warnf("Failed to find patient profile: %+v", err)
		return nil, err
	}
	if profile == nil {
		return nil, ErrPatientNotFound
	}

	user, err := u.userRepo.FindByID(tx, userID)
	if err != nil {
		u.log.Warnf("Failed to find user: %+v", err)
		return nil, err
	}

	// Capture old value for audit
	oldValue := converter.PatientProfileToResponse(profile, user)

	// Update allowed fields
	updated := false

	if req.Password != "" {
		// Validate old password
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
			return nil, ErrInvalidOldPassword
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			u.log.Warnf("Failed to hash password: %+v", err)
			return nil, err
		}
		user.Password = string(hashedPassword)
		updated = true
	}

	if req.PhoneNumber != "" {
		profile.PhoneNumber = req.PhoneNumber
		updated = true
	}

	if req.Address != "" {
		profile.Address = req.Address
		updated = true
	}

	if !updated {
		return converter.PatientProfileToResponse(profile, user), nil
	}

	// Update user (for password change)
	if err := u.userRepo.Update(tx, user); err != nil {
		u.log.Warnf("Failed to update user: %+v", err)
		return nil, err
	}

	// Update patient profile (for phone_number, address)
	if err := u.patientProfileRepo.Update(ctx, tx, profile); err != nil {
		u.log.Warnf("Failed to update patient profile: %+v", err)
		return nil, err
	}

	// Audit log
	newValue := converter.PatientProfileToResponse(profile, user)
	if err := u.auditService.LogUpdate(ctx, tx, &userID, entity.AuditActionProfileUpdate, "patient_profile", userID.String(), oldValue, newValue); err != nil {
		u.log.Warnf("Failed to create audit log: %+v", err)
	}

	if err := tx.Commit().Error; err != nil {
		u.log.Warnf("Failed commit transaction: %+v", err)
		return nil, err
	}

	return converter.PatientProfileToResponse(profile, user), nil
}
