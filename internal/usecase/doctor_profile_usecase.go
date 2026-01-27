package usecase

import (
	"context"
	"errors"

	"go-template-clean-architecture/internal/converter"
	"go-template-clean-architecture/internal/delivery/dto"
	"go-template-clean-architecture/internal/domain/entity"
	"go-template-clean-architecture/internal/domain/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrDoctorNotFound     = errors.New("doctor not found")
	ErrDoctorEmailExists  = errors.New("email already exists")
	ErrDoctorSTRExists    = errors.New("STR number already exists")
	ErrDoctorRoleNotFound = errors.New("role not found")
)

type DoctorProfileUsecase interface {
	CreateDoctor(ctx context.Context, req *dto.CreateDoctorRequest) (*dto.DoctorResponse, error)
	GetDoctor(ctx context.Context, doctorID uuid.UUID) (*dto.DoctorResponse, error)
	GetAllDoctors(ctx context.Context) (*dto.DoctorListResponse, error)
	UpdateDoctor(ctx context.Context, doctorID uuid.UUID, req *dto.UpdateDoctorRequest) (*dto.DoctorResponse, error)
	DeleteDoctor(ctx context.Context, doctorID uuid.UUID) error
}

type doctorProfileUsecase struct {
	db                *gorm.DB
	log               *logrus.Logger
	userRepo          repository.UserRepository
	doctorProfileRepo repository.DoctorProfileRepository
}

func NewDoctorProfileUsecase(
	db *gorm.DB,
	log *logrus.Logger,
	userRepo repository.UserRepository,
	doctorProfileRepo repository.DoctorProfileRepository,
) DoctorProfileUsecase {
	return &doctorProfileUsecase{
		db:                db,
		log:               log,
		userRepo:          userRepo,
		doctorProfileRepo: doctorProfileRepo,
	}
}

func (u *doctorProfileUsecase) CreateDoctor(ctx context.Context, req *dto.CreateDoctorRequest) (*dto.DoctorResponse, error) {
	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		u.log.Warnf("Failed to hash password: %+v", err)
		return nil, err
	}

	// Create user with doctor profile in single insert using GORM association
	doctorProfile := &entity.DoctorProfile{
		STRNumber:      req.STRNumber,
		Specialization: req.Specialization,
		Biography:      req.Biography,
		User: entity.User{
			Email:    req.Email,
			Password: string(hashedPassword),
			FullName: req.FullName,
			RoleID:   entity.RoleIDDoctor,
		},
	}
	if err := u.doctorProfileRepo.Create(tx, doctorProfile); err != nil {
		u.log.Warnf("Failed to create doctor: %+v", err)
		if isDuplicateKeyError(err, "email") {
			return nil, ErrDoctorEmailExists
		}
		if isDuplicateKeyError(err, "str_number") {
			return nil, ErrDoctorSTRExists
		}
		if isForeignKeyError(err, "role") {
			return nil, ErrDoctorRoleNotFound
		}
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		u.log.Warnf("Failed commit transaction: %+v", err)
		return nil, err
	}

	return converter.DoctorProfileToResponse(doctorProfile), nil
}

func (u *doctorProfileUsecase) GetDoctor(ctx context.Context, userID uuid.UUID) (*dto.DoctorResponse, error) {
	profile, err := u.doctorProfileRepo.FindByUserID(u.db, userID)
	if err != nil {
		u.log.Warnf("Failed to find doctor profile: %+v", err)
		return nil, err
	}
	if profile == nil {
		u.log.Warnf("Failed to find doctor profile: %+v", "doctor not found")
		return nil, ErrDoctorNotFound
	}

	return converter.DoctorProfileToResponse(profile), nil
}

func (u *doctorProfileUsecase) GetAllDoctors(ctx context.Context) (*dto.DoctorListResponse, error) {
	profiles, err := u.doctorProfileRepo.FindAll(u.db)
	if err != nil {
		u.log.Warnf("Failed to find all doctor profiles: %+v", err)
		return nil, err
	}

	doctors := converter.DoctorProfilesToResponses(profiles)

	return &dto.DoctorListResponse{
		Doctors: doctors,
		Total:   len(doctors),
	}, nil
}

func (u *doctorProfileUsecase) UpdateDoctor(ctx context.Context, userID uuid.UUID, req *dto.UpdateDoctorRequest) (*dto.DoctorResponse, error) {
	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	// get doctor profile
	profile, err := u.doctorProfileRepo.FindByUserID(tx, userID)
	if err != nil {
		u.log.Warnf("Failed to find doctor profile: %+v", err)
		return nil, err
	}

	if profile == nil {
		u.log.Warnf("Failed to find doctor profile: %+v", "doctor not found")
		return nil, ErrDoctorNotFound
	}

	// set doctor profile & user
	if req.Email != "" {
		profile.User.Email = req.Email
	}
	if req.Password != "" {
		profile.User.Password = req.Password
	}
	if req.FullName != "" {
		profile.User.FullName = req.FullName
	}
	if req.IsActive != nil {
		profile.User.IsActive = req.IsActive
	}
	if req.STRNumber != "" {
		profile.STRNumber = req.STRNumber
	}
	if req.Specialization != "" {
		profile.Specialization = req.Specialization
	}
	if req.Biography != "" {
		profile.Biography = req.Biography
	}

	// Update profile
	if err := u.doctorProfileRepo.Update(tx, profile); err != nil {
		if isDuplicateKeyError(err, "str_number") {
			return nil, ErrDoctorSTRExists
		}
		u.log.Warnf("Failed to update doctor profile: %+v", err)
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		u.log.Warnf("Failed commit transaction: %+v", err)
		return nil, err
	}

	return converter.DoctorProfileToResponse(profile), nil
}

func (u *doctorProfileUsecase) DeleteDoctor(ctx context.Context, userID uuid.UUID) error {
	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	profile, err := u.doctorProfileRepo.FindByUserID(tx, userID)
	if err != nil {
		u.log.Warnf("Failed to find doctor profile: %+v", err)
		return err
	}
	if profile == nil {
		return ErrDoctorNotFound
	}

	// Delete doctor profile first (foreign key constraint)
	if err := u.doctorProfileRepo.Delete(tx, userID); err != nil {
		u.log.Warnf("Failed to delete doctor profile: %+v", err)
		return err
	}

	// Delete user
	if err := tx.Where("id = ?", userID).Delete(&entity.User{}).Error; err != nil {
		u.log.Warnf("Failed to delete user: %+v", err)
		return err
	}

	if err := tx.Commit().Error; err != nil {
		u.log.Warnf("Failed commit transaction: %+v", err)
		return err
	}

	return nil
}
