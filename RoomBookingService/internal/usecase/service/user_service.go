package service

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
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, req dto.CreateUserRequest) (*dto.AuthResult, error)
	GetUserByEmail(ctx context.Context, email string) (*dto.AuthResult, error)
	DummyLogin(ctx context.Context, role string) (*dto.TokenResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.TokenResponse, error)
	RegisterWithPassword(ctx context.Context, req dto.CreateUserRequest) (*dto.AuthResult, error)
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

	result = &dto.AuthResult{
		Email: req.Email,
		Role:  req.Role,
	}

	return result, err
}

func (s *authService) GetUserByEmail(ctx context.Context, email string) (*dto.AuthResult, error) {
	log := logger.FromContext(ctx)

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		log.Error("failed to get user by email", "email", email, "error", err)
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
		return nil, myerrors.ErrInvalidRole
	}

	token, err := s.tokenService.GenerateDummyToken(role)
	if err != nil {
		log.Error("dummy login: failed to generate token", "role", role, "error", err)
		return nil, err
	}

	return &dto.TokenResponse{Token: token}, nil
}

func (s *authService) RegisterWithPassword(ctx context.Context, req dto.CreateUserRequest) (*dto.AuthResult, error) {
	log := logger.FromContext(ctx)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if err := s.validate.Struct(req); err != nil {
		log.Warn("auth service: register validation failed", "email", req.Email, "error", err)
		return nil, err
	}
	if req.Password == "" {
		return nil, myerrors.ErrInvalidRequest
	}
	var result *dto.AuthResult

	_, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, myerrors.ErrUserAlreadyExists
	}
	if !errors.Is(err, myerrors.ErrUserNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:    req.Email,
		Role:     req.Role,
		Password: string(hash),
	}
	err = s.userRepo.RegisterWithPassword(ctx, user)
	if err != nil {
		return nil, err
	}

	result = &dto.AuthResult{
		UserID: user.ID,
		Email:  req.Email,
		Role:   req.Role,
	}

	return result, nil
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.TokenResponse, error) {
	log := logger.FromContext(ctx)

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if err := s.validate.Struct(req); err != nil {
		log.Warn("auth service: login validation failed", "email", req.Email, "password_len", len(req.Password), "error", err)
		return nil, err
	}

	log.Info("auth service: login attempt", "email", req.Email)
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Warn("auth service: login lookup failed", "email", req.Email, "error", err)
		return nil, myerrors.ErrInvalidCredentials
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		log.Warn("auth service: login password mismatch", "email", req.Email)
		return nil, myerrors.ErrInvalidCredentials
	}
	token, err := s.tokenService.GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	result := &dto.TokenResponse{
		Token: token,
	}
	return result, nil
}
