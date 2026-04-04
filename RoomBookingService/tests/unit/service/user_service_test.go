package unit_test

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/model"
	"RoomBookingService/internal/usecase/service"
	myerrors "RoomBookingService/pkg/my_errors"
	"context"
	"errors"
	"testing"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
)

type userRepoStub struct {
	getByEmailResult *model.User
	getByEmailErr    error
}

func (u *userRepoStub) GetUserByEmail(_ context.Context, _ string) (*model.User, error) {
	if u.getByEmailErr != nil {
		return nil, u.getByEmailErr
	}
	return u.getByEmailResult, nil
}

func (u *userRepoStub) RegisterWithPassword(_ context.Context, user *model.User) error {
	if user != nil && user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	return nil
}

type tokenServiceStub struct {
	token string
	err   error
}

func (t *tokenServiceStub) GenerateDummyToken(_ string) (string, error) {
	if t.err != nil {
		return "", t.err
	}
	return t.token, nil
}

func (t *tokenServiceStub) ParseToken(_ string) (uuid.UUID, string, error) {
	return uuid.Nil, "", errors.New("not used")
}

func (t *tokenServiceStub) GenerateToken(_ uuid.UUID, _ string) (string, error) {
	if t.err != nil {
		return "", t.err
	}
	return t.token, nil
}

func TestAuthService_DummyLogin_Success(t *testing.T) {
	svc := service.NewUserService(&userRepoStub{}, validator.New(), &tokenServiceStub{token: "abc"})

	res, err := svc.DummyLogin(context.Background(), "user")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Token != "abc" {
		t.Fatalf("expected token abc, got %q", res.Token)
	}
}

func TestAuthService_DummyLogin_InvalidRole(t *testing.T) {
	svc := service.NewUserService(&userRepoStub{}, validator.New(), &tokenServiceStub{token: "abc"})

	_, err := svc.DummyLogin(context.Background(), "manager")
	if !errors.Is(err, myerrors.ErrInvalidRole) {
		t.Fatalf("expected ErrInvalidRole, got %v", err)
	}
}

func TestAuthService_GetUserByEmail_Success(t *testing.T) {
	u := &model.User{ID: uuid.New(), Email: "a@b.com", Role: "user"}
	svc := service.NewUserService(&userRepoStub{getByEmailResult: u}, validator.New(), &tokenServiceStub{token: "abc"})

	res, err := svc.GetUserByEmail(context.Background(), "a@b.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Email != "a@b.com" {
		t.Fatalf("expected email a@b.com, got %q", res.Email)
	}
}

func TestAuthService_Register_AlreadyExists(t *testing.T) {
	svc := service.NewUserService(&userRepoStub{getByEmailResult: &model.User{ID: uuid.New()}}, validator.New(), &tokenServiceStub{token: "abc"})

	_, err := svc.Register(context.Background(), dto.CreateUserRequest{Email: "u@e.com", Role: "user"})
	if !errors.Is(err, myerrors.ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
}
