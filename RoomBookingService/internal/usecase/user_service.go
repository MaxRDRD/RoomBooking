package usecase

import (
	"RoomBookingService/internal/auth"
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/logger"
	"RoomBookingService/internal/model"
	"RoomBookingService/internal/repository"
	myerrors "RoomBookingService/pkg/my_errors"

	"context"
	"errors"
	"strings"

	"github.com/go-playground/validator"
)

type AuthService interface {
	Register(ctx context.Context, req dto.CreateUserRequest) (*dto.AuthResult, error)
	GetUserByEmail(ctx context.Context, email string) (*dto.AuthResult, error)
	DummyLogin(ctx context.Context, role string) (*dto.TokenResponse, error)
}

type authService struct {
	userRepo     repository.UserRepository
	validate     *validator.Validate
	tokenService auth.TokenService
}

func NewUserService(userRepo repository.UserRepository,
	validator *validator.Validate,
	tokenService auth.TokenService,
) AuthService {
	return &authService{
		userRepo:     userRepo,
		validate:     validator,
		tokenService: tokenService,
	}
}

func (s *authService) Register(ctx context.Context, req dto.CreateUserRequest) (*dto.AuthResult, error) {
	log := logger.FromContext(ctx)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if err := s.validate.Struct(req); err != nil {
		log.Warn("auth service: register validation failed", "email", req.Email, "error", err)
		return nil, err
	}
	var result *dto.AuthResult

	_, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, myerrors.ErrUserAlreadyExists
	}
	if !errors.Is(err, myerrors.ErrUserNotFound) {
		return nil, err
	}

	user := &model.User{
		Email: req.Email,
		Role:  req.Role,
	}
	result = &dto.AuthResult{
		Email: user.Email,
		Role:  user.Role,
	}

	return result, err
}

func (s *authService) GetUserByEmail(ctx context.Context, email string) (*dto.AuthResult, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &dto.AuthResult{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
	}, nil
}

func (s *authService) DummyLogin(ctx context.Context, role string) (*dto.TokenResponse, error) {
	log := logger.FromContext(ctx)
	role = strings.ToLower(strings.TrimSpace(role))

	if role != "admin" && role != "user" {
		log.Warn("dummy login: invalid role", "role", role)
		return nil, errors.New("invalid role")
	}

	token, err := s.tokenService.GenerateDummyToken(role)
	if err != nil {
		log.Error("dummy login: failed to generate token", "role", role, "error", err)
		return nil, err
	}

	return &dto.TokenResponse{Token: token}, nil
}
