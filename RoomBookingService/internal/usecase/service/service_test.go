package service

import (
	"RoomBookingService/internal/auth"
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/model"
	"RoomBookingService/internal/repository"
	myerrors "RoomBookingService/pkg/my_errors"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type userRepoStub struct {
	getUserByEmailFn       func(context.Context, string) (*model.User, error)
	registerWithPasswordFn func(context.Context, *model.User) error
}

func (u *userRepoStub) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	if u.getUserByEmailFn != nil {
		return u.getUserByEmailFn(ctx, email)
	}
	return nil, myerrors.ErrUserNotFound
}

func (u *userRepoStub) RegisterWithPassword(ctx context.Context, user *model.User) error {
	if u.registerWithPasswordFn != nil {
		return u.registerWithPasswordFn(ctx, user)
	}
	user.ID = uuid.New()
	return nil
}

type roomRepoStub struct {
	createFn func(context.Context, *model.Room) error
	getAllFn func(context.Context) ([]*model.Room, error)
}

func (r *roomRepoStub) CreateRoom(ctx context.Context, room *model.Room) error {
	if r.createFn != nil {
		return r.createFn(ctx, room)
	}
	room.ID = uuid.New()
	return nil
}

func (r *roomRepoStub) GetAllRooms(ctx context.Context) ([]*model.Room, error) {
	if r.getAllFn != nil {
		return r.getAllFn(ctx)
	}
	return []*model.Room{}, nil
}

type scheduleRepoStub struct {
	createFn func(context.Context, *model.Schedule) error
}

func (s *scheduleRepoStub) CreateSchedule(ctx context.Context, schedule *model.Schedule) error {
	if s.createFn != nil {
		return s.createFn(ctx, schedule)
	}
	schedule.ID = uuid.New()
	return nil
}

type bookingRepoStub struct {
	createFn func(context.Context, uuid.UUID, dto.BookingRequest) (model.Booking, error)
	cancelFn func(context.Context, uuid.UUID, dto.CancelBookingRequest) (model.Booking, error)
	getMyFn  func(context.Context, uuid.UUID) ([]model.Booking, error)
	getAllFn func(context.Context, dto.BookingFilter) ([]model.Booking, int, error)
}

func (b *bookingRepoStub) CreateBooking(ctx context.Context, userID uuid.UUID, req dto.BookingRequest) (model.Booking, error) {
	if b.createFn != nil {
		return b.createFn(ctx, userID, req)
	}
	return model.Booking{ID: uuid.New(), UserID: userID, SlotID: req.SlotID, Status: "active"}, nil
}

func (b *bookingRepoStub) CancelBooking(ctx context.Context, userID uuid.UUID, req dto.CancelBookingRequest) (model.Booking, error) {
	if b.cancelFn != nil {
		return b.cancelFn(ctx, userID, req)
	}
	return model.Booking{ID: req.BookingID, UserID: userID, Status: "cancelled"}, nil
}

func (b *bookingRepoStub) GetMyBookings(ctx context.Context, userID uuid.UUID) ([]model.Booking, error) {
	if b.getMyFn != nil {
		return b.getMyFn(ctx, userID)
	}
	return []model.Booking{}, nil
}

func (b *bookingRepoStub) GetAllBookings(ctx context.Context, filter dto.BookingFilter) ([]model.Booking, int, error) {
	if b.getAllFn != nil {
		return b.getAllFn(ctx, filter)
	}
	return []model.Booking{}, 0, nil
}

type slotRepoStub struct {
	getFn func(context.Context, dto.SlotsRequest) ([]model.Slot, error)
}

func (s *slotRepoStub) GetAvailableSlots(ctx context.Context, req dto.SlotsRequest) ([]model.Slot, error) {
	if s.getFn != nil {
		return s.getFn(ctx, req)
	}
	return []model.Slot{}, nil
}

type tokenServiceStub struct {
	generateFn      func(string) (string, error)
	generateTokenFn func(uuid.UUID, string) (string, error)
	parseFn         func(string) (uuid.UUID, string, error)
}

func (t *tokenServiceStub) GenerateDummyToken(role string) (string, error) {
	if t.generateFn != nil {
		return t.generateFn(role)
	}
	return "token-" + role, nil
}

func (t *tokenServiceStub) ParseToken(tokenString string) (uuid.UUID, string, error) {
	if t.parseFn != nil {
		return t.parseFn(tokenString)
	}
	return uuid.Nil, "", nil
}

func (t *tokenServiceStub) GenerateToken(userID uuid.UUID, role string) (string, error) {
	if t.generateTokenFn != nil {
		return t.generateTokenFn(userID, role)
	}
	return "token", nil
}

func TestParseTimeOfDay(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		parsed, err := parseTimeOfDay("09:30")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if parsed.Hour() != 9 || parsed.Minute() != 30 {
			t.Fatalf("unexpected time parsed: %v", parsed)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := parseTimeOfDay("25:00")
		if !errors.Is(err, myerrors.ErrInvalidRequest) {
			t.Fatalf("expected ErrInvalidRequest, got %v", err)
		}
	})
}

func TestAuthService_DummyLogin(t *testing.T) {
	svc := NewUserService(&userRepoStub{}, validator.New(), &tokenServiceStub{generateFn: func(role string) (string, error) {
		return "token-" + role, nil
	}})

	res, err := svc.DummyLogin(context.Background(), " admin ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Token != "token-admin" {
		t.Fatalf("unexpected token %q", res.Token)
	}
}

func TestAuthService_DummyLogin_InvalidRole(t *testing.T) {
	svc := NewUserService(&userRepoStub{}, validator.New(), &tokenServiceStub{})
	_, err := svc.DummyLogin(context.Background(), "manager")
	if !errors.Is(err, myerrors.ErrInvalidRole) {
		t.Fatalf("expected ErrInvalidRole, got %v", err)
	}
}

func TestAuthService_GetUserByEmail(t *testing.T) {
	expected := &model.User{ID: uuid.New(), Email: "u@example.com", Role: "user"}
	svc := NewUserService(&userRepoStub{getUserByEmailFn: func(context.Context, string) (*model.User, error) {
		return expected, nil
	}}, validator.New(), &tokenServiceStub{})

	res, err := svc.GetUserByEmail(context.Background(), expected.Email)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.UserID != expected.ID {
		t.Fatalf("unexpected user id %v", res.UserID)
	}
}

func TestAuthService_Register(t *testing.T) {
	svc := NewUserService(&userRepoStub{getUserByEmailFn: func(context.Context, string) (*model.User, error) {
		return nil, myerrors.ErrUserNotFound
	}}, validator.New(), &tokenServiceStub{})

	res, err := svc.Register(context.Background(), dto.CreateUserRequest{Email: "User@Example.com", Role: "user"})
	if !errors.Is(err, myerrors.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil result")
	}
	if res.Email != "user@example.com" {
		t.Fatalf("expected normalized email, got %q", res.Email)
	}
}

func TestAuthService_RegisterAlreadyExists(t *testing.T) {
	svc := NewUserService(&userRepoStub{getUserByEmailFn: func(context.Context, string) (*model.User, error) {
		return &model.User{ID: uuid.New()}, nil
	}}, validator.New(), &tokenServiceStub{})

	_, err := svc.Register(context.Background(), dto.CreateUserRequest{Email: "user@example.com", Role: "user"})
	if !errors.Is(err, myerrors.ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestAuthService_RegisterWithPassword(t *testing.T) {
	var stored *model.User
	svc := NewUserService(&userRepoStub{
		getUserByEmailFn: func(context.Context, string) (*model.User, error) {
			return nil, myerrors.ErrUserNotFound
		},
		registerWithPasswordFn: func(_ context.Context, user *model.User) error {
			stored = user
			user.ID = uuid.New()
			return nil
		},
	}, validator.New(), &tokenServiceStub{})

	res, err := svc.RegisterWithPassword(context.Background(), dto.CreateUserRequest{
		Email:    "User@Example.com",
		Password: "secret12",
		Role:     "user",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.UserID == uuid.Nil || res.Email != "user@example.com" {
		t.Fatalf("unexpected register result: %+v", res)
	}
	if stored == nil {
		t.Fatal("expected user to be persisted")
	}
	if stored.Password == "secret12" {
		t.Fatal("expected password hash, got plaintext")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte("secret12")); err != nil {
		t.Fatalf("stored hash mismatch: %v", err)
	}
}

func TestAuthService_Login(t *testing.T) {
	userID := uuid.New()
	hash, err := bcrypt.GenerateFromPassword([]byte("secret12"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}

	svc := NewUserService(&userRepoStub{
		getUserByEmailFn: func(context.Context, string) (*model.User, error) {
			return &model.User{ID: userID, Email: "user@example.com", Role: "user", Password: string(hash)}, nil
		},
	}, validator.New(), &tokenServiceStub{
		generateTokenFn: func(id uuid.UUID, role string) (string, error) {
			if id != userID || role != "user" {
				t.Fatalf("unexpected token payload id=%s role=%s", id, role)
			}
			return "jwt-token", nil
		},
	})

	res, err := svc.Login(context.Background(), dto.LoginRequest{Email: "user@example.com", Password: "secret12"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Token != "jwt-token" {
		t.Fatalf("unexpected token %q", res.Token)
	}
}

func TestRoomService(t *testing.T) {
	repo := &roomRepoStub{}
	svc := NewRoomService(repo)
	cap := 6

	room, err := svc.CreateRoom(context.Background(), dto.CreateRoomRequest{Name: "A101", Description: "main room", Capacity: &cap})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if room.Description != "main room" || room.Capacity != 6 {
		t.Fatalf("unexpected room payload: %+v", room)
	}

	rooms, err := svc.GetAllRooms(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(rooms) != 0 {
		t.Fatalf("expected empty room list from stub, got %d", len(rooms))
	}
}

func TestScheduleService(t *testing.T) {
	repo := &scheduleRepoStub{}
	svc := NewScheduleService(repo)

	schedule, err := svc.CreateSchedule(context.Background(), dto.CreateScheduleRequest{
		RoomID:     uuid.New(),
		DaysOfWeek: []int{1, 3, 5},
		StartTime:  "09:00",
		EndTime:    "18:00",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if schedule.StartTime.Hour() != 9 || schedule.EndTime.Hour() != 18 {
		t.Fatalf("unexpected schedule times: %+v", schedule)
	}

	_, err = svc.CreateSchedule(context.Background(), dto.CreateScheduleRequest{
		RoomID:     uuid.New(),
		DaysOfWeek: []int{1},
		StartTime:  "25:00",
		EndTime:    "26:00",
	})
	if !errors.Is(err, myerrors.ErrInvalidRequest) {
		t.Fatalf("expected ErrInvalidRequest, got %v", err)
	}

	_, err = svc.CreateSchedule(context.Background(), dto.CreateScheduleRequest{
		RoomID:     uuid.New(),
		DaysOfWeek: []int{1},
		StartTime:  "09:00",
		EndTime:    "10:00",
	})
	if err != nil {
		_ = err
	}

	_, err = (&scheduleService{scheduleRepo: &scheduleRepoStub{createFn: func(context.Context, *model.Schedule) error {
		return &pgconn.PgError{Code: "23505"}
	}}}).CreateSchedule(context.Background(), dto.CreateScheduleRequest{
		RoomID:     uuid.New(),
		DaysOfWeek: []int{1},
		StartTime:  "09:00",
		EndTime:    "10:00",
	})
	if !errors.Is(err, myerrors.ErrScheduleAlreadyExists) {
		t.Fatalf("expected ErrScheduleAlreadyExists, got %v", err)
	}
}

func TestBookingService(t *testing.T) {
	repo := &bookingRepoStub{}
	svc := NewBookingService(repo)
	userID := uuid.New()
	slotID := uuid.New()
	bookingID := uuid.New()

	created, err := svc.CreateBooking(context.Background(), userID, dto.BookingRequest{SlotID: slotID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created.UserID != userID || created.SlotID != slotID {
		t.Fatalf("unexpected booking payload: %+v", created)
	}

	cancelled, err := svc.CancelBooking(context.Background(), userID, dto.CancelBookingRequest{BookingID: bookingID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cancelled.Status != "cancelled" {
		t.Fatalf("unexpected booking status: %s", cancelled.Status)
	}

	myBookings, err := svc.GetMyBookings(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(myBookings) != 0 {
		t.Fatalf("expected empty list, got %d", len(myBookings))
	}

	allBookings, total, err := svc.GetAllBookings(context.Background(), dto.BookingFilter{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(allBookings) != 0 || total != 0 {
		t.Fatalf("unexpected admin bookings payload: %d/%d", len(allBookings), total)
	}
}

func TestSlotService(t *testing.T) {
	repo := &slotRepoStub{}
	svc := NewSlotService(repo)

	slots, err := svc.GetAvailableSlots(context.Background(), dto.SlotsRequest{RoomID: uuid.New(), Date: time.Now().UTC()})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(slots) != 0 {
		t.Fatalf("expected empty slots list, got %d", len(slots))
	}
}

var _ repository.UserRepository = (*userRepoStub)(nil)
var _ repository.RoomRepository = (*roomRepoStub)(nil)
var _ repository.ScheduleRepository = (*scheduleRepoStub)(nil)
var _ repository.BookingRepository = (*bookingRepoStub)(nil)
var _ repository.SlotRepository = (*slotRepoStub)(nil)
var _ auth.TokenService = (*tokenServiceStub)(nil)
