package handler

import (
	"RoomBookingService/internal/auth"
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/model"
	"RoomBookingService/internal/repository"
	"RoomBookingService/internal/usecase/service"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
)

type authServiceStub struct {
	dummyLoginFn           func(context.Context, string) (*dto.TokenResponse, error)
	registerFn             func(context.Context, dto.CreateUserRequest) (*dto.AuthResult, error)
	getUserByEmailFn       func(context.Context, string) (*dto.AuthResult, error)
	registerWithPasswordFn func(context.Context, dto.CreateUserRequest) (*dto.AuthResult, error)
	loginFn                func(context.Context, dto.LoginRequest) (*dto.TokenResponse, error)
}

func (a *authServiceStub) Register(ctx context.Context, req dto.CreateUserRequest) (*dto.AuthResult, error) {
	if a.registerFn != nil {
		return a.registerFn(ctx, req)
	}
	return &dto.AuthResult{Email: req.Email, Role: req.Role}, nil
}

func (a *authServiceStub) GetUserByEmail(ctx context.Context, email string) (*dto.AuthResult, error) {
	if a.getUserByEmailFn != nil {
		return a.getUserByEmailFn(ctx, email)
	}
	return &dto.AuthResult{Email: email, Role: "user"}, nil
}

func (a *authServiceStub) DummyLogin(ctx context.Context, role string) (*dto.TokenResponse, error) {
	if a.dummyLoginFn != nil {
		return a.dummyLoginFn(ctx, role)
	}
	return &dto.TokenResponse{Token: "dummy-token-" + role}, nil
}

func (a *authServiceStub) Login(ctx context.Context, req dto.LoginRequest) (*dto.TokenResponse, error) {
	if a.loginFn != nil {
		return a.loginFn(ctx, req)
	}
	return &dto.TokenResponse{Token: "token"}, nil
}

func (a *authServiceStub) RegisterWithPassword(ctx context.Context, req dto.CreateUserRequest) (*dto.AuthResult, error) {
	if a.registerWithPasswordFn != nil {
		return a.registerWithPasswordFn(ctx, req)
	}
	return &dto.AuthResult{Email: req.Email, Role: req.Role}, nil
}

type roomServiceStub struct {
	createFn func(context.Context, dto.CreateRoomRequest) (*model.Room, error)
	getAllFn func(context.Context) ([]*model.Room, error)
}

func (r *roomServiceStub) CreateRoom(ctx context.Context, req dto.CreateRoomRequest) (*model.Room, error) {
	if r.createFn != nil {
		return r.createFn(ctx, req)
	}
	return &model.Room{ID: uuid.New(), Name: req.Name, Description: req.Description}, nil
}

func (r *roomServiceStub) GetAllRooms(ctx context.Context) ([]*model.Room, error) {
	if r.getAllFn != nil {
		return r.getAllFn(ctx)
	}
	return []*model.Room{{ID: uuid.New(), Name: "A101"}}, nil
}

type scheduleServiceStub struct {
	createFn func(context.Context, dto.CreateScheduleRequest) (*model.Schedule, error)
}

func (s *scheduleServiceStub) CreateSchedule(ctx context.Context, req dto.CreateScheduleRequest) (*model.Schedule, error) {
	if s.createFn != nil {
		return s.createFn(ctx, req)
	}
	return &model.Schedule{ID: uuid.New(), RoomID: req.RoomID}, nil
}

type slotServiceStub struct {
	getFn func(context.Context, dto.SlotsRequest) ([]model.Slot, error)
}

func (s *slotServiceStub) GetAvailableSlots(ctx context.Context, req dto.SlotsRequest) ([]model.Slot, error) {
	if s.getFn != nil {
		return s.getFn(ctx, req)
	}
	return []model.Slot{{ID: uuid.New(), RoomID: req.RoomID, StartAt: req.Date, EndAt: req.Date.Add(30 * time.Minute)}}, nil
}

type bookingServiceStub struct {
	createFn func(context.Context, uuid.UUID, dto.BookingRequest) (model.Booking, error)
	cancelFn func(context.Context, uuid.UUID, dto.CancelBookingRequest) (model.Booking, error)
	myFn     func(context.Context, uuid.UUID) ([]model.Booking, error)
	allFn    func(context.Context, dto.BookingFilter) ([]model.Booking, int, error)
}

func (b *bookingServiceStub) CreateBooking(ctx context.Context, userID uuid.UUID, req dto.BookingRequest) (model.Booking, error) {
	if b.createFn != nil {
		return b.createFn(ctx, userID, req)
	}
	return model.Booking{ID: uuid.New(), UserID: userID, SlotID: req.SlotID, Status: "active"}, nil
}

func (b *bookingServiceStub) CancelBooking(ctx context.Context, userID uuid.UUID, req dto.CancelBookingRequest) (model.Booking, error) {
	if b.cancelFn != nil {
		return b.cancelFn(ctx, userID, req)
	}
	return model.Booking{ID: req.BookingID, UserID: userID, Status: "cancelled"}, nil
}

func (b *bookingServiceStub) GetMyBookings(ctx context.Context, userID uuid.UUID) ([]model.Booking, error) {
	if b.myFn != nil {
		return b.myFn(ctx, userID)
	}
	return []model.Booking{{ID: uuid.New(), UserID: userID, Status: "active"}}, nil
}

func (b *bookingServiceStub) GetAllBookings(ctx context.Context, filter dto.BookingFilter) ([]model.Booking, int, error) {
	if b.allFn != nil {
		return b.allFn(ctx, filter)
	}
	return []model.Booking{{ID: uuid.New()}}, 1, nil
}

func TestUserHandler_DummyLogin(t *testing.T) {
	h := NewUserHandler(&authServiceStub{})
	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBufferString(`{"role":"user"}`))
	rec := httptest.NewRecorder()
	h.DummyLogin(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestUserHandler_RegisterWithPassword(t *testing.T) {
	h := NewUserHandler(&authServiceStub{})
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(`{"email":"user@example.com","password":"secret12","role":"user"}`))
	rec := httptest.NewRecorder()
	h.RegisterUserWithPassword(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestUserHandler_Login(t *testing.T) {
	h := NewUserHandler(&authServiceStub{})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":"user@example.com","password":"secret12"}`))
	rec := httptest.NewRecorder()
	h.Login(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRoomHandler_RegisterAndList(t *testing.T) {
	h := NewRoomHandler(&roomServiceStub{})
	createReq := httptest.NewRequest(http.MethodPost, "/rooms/create", bytes.NewBufferString(`{"name":"A101","description":"main","capacity":6}`))
	createRec := httptest.NewRecorder()
	h.RegisterRoomRoutes(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", createRec.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/rooms/list", nil)
	listRec := httptest.NewRecorder()
	h.GetAllRooms(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", listRec.Code)
	}
}

func TestScheduleHandler_CreateSchedule(t *testing.T) {
	h := NewScheduleHandler(&scheduleServiceStub{}, validator.New())
	req := httptest.NewRequest(http.MethodPost, "/rooms/invalid/schedule/create", bytes.NewBufferString(`{"daysOfWeek":[1],"startTime":"09:00","endTime":"10:00"}`))
	rec := httptest.NewRecorder()
	h.CreateSchedule(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestSlotHandler_GetAvailableSlots(t *testing.T) {
	slotService := service.NewSlotService(&slotServiceStub{})
	h := NewSlotHandler(*slotService)
	req := httptest.NewRequest(http.MethodGet, "/rooms/"+uuid.NewString()+"/slots/list?date=2026-04-01", nil)
	rec := httptest.NewRecorder()
	h.GetAvailableSlots(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestBookingHandler_CreateAndCancel(t *testing.T) {
	bookingSvc := &bookingServiceStub{}
	h := NewBookingHandler(bookingSvc, validator.New())
	jwtService := auth.NewJWTService()
	token, err := jwtService.GenerateDummyToken("user")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	createReq := httptest.NewRequest(http.MethodPost, "/bookings/create", bytes.NewBufferString(`{"slotId":"11111111-1111-1111-1111-111111111111"}`))
	createReq.Header.Set("Authorization", "Bearer "+token)
	createRec := httptest.NewRecorder()
	auth.NewAuthMiddleware(jwtService).JWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.CreateBooking(w, r)
	})).ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", createRec.Code)
	}

	cancelReq := httptest.NewRequest(http.MethodPost, "/bookings/"+uuid.NewString()+"/cancel", nil)
	cancelReq.Header.Set("Authorization", "Bearer "+token)
	cancelRec := httptest.NewRecorder()
	auth.NewAuthMiddleware(jwtService).JWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.CancelBooking(w, r)
	})).ServeHTTP(cancelRec, cancelReq)
	if cancelRec.Code != http.StatusBadRequest && cancelRec.Code != http.StatusOK {
		t.Fatalf("unexpected cancel code %d", cancelRec.Code)
	}
}

var _ service.AuthService = (*authServiceStub)(nil)
var _ service.RoomService = (*roomServiceStub)(nil)
var _ service.ScheduleService = (*scheduleServiceStub)(nil)
var _ repository.SlotRepository = (*slotServiceStub)(nil)
var _ service.BookingService = (*bookingServiceStub)(nil)
