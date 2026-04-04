package e2e_test

import (
	"RoomBookingService/cmd/server"
	"RoomBookingService/internal/auth"
	"RoomBookingService/internal/handler"
	"RoomBookingService/internal/repository_impl/postgres"
	"RoomBookingService/internal/usecase/service"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator"
	"github.com/jackc/pgx/v5/pgxpool"
)

const sharedTestDBLockID int64 = 482913741

func TestE2E_CreateRoomScheduleBooking(t *testing.T) {
	testEnv := newE2EEnv(t)
	defer testEnv.close()

	adminToken := testEnv.dummyLogin(t, "admin")
	userToken := testEnv.dummyLogin(t, "user")

	roomID := testEnv.createRoom(t, adminToken)

	targetDate := time.Now().UTC().Add(24 * time.Hour)
	day := isoWeekday(targetDate)
	testEnv.createSchedule(t, adminToken, roomID, []int{day}, "09:00", "11:00")

	slotID := testEnv.getFirstAvailableSlot(t, userToken, roomID, targetDate.Format("2006-01-02"))
	bookingID, status := testEnv.createBooking(t, userToken, slotID)

	if bookingID == "" {
		t.Fatalf("expected booking id to be set")
	}
	if status != "active" {
		t.Fatalf("expected booking status=active, got %q", status)
	}
}

func TestE2E_CancelBooking_Idempotent(t *testing.T) {
	testEnv := newE2EEnv(t)
	defer testEnv.close()

	adminToken := testEnv.dummyLogin(t, "admin")
	userToken := testEnv.dummyLogin(t, "user")

	roomID := testEnv.createRoom(t, adminToken)
	targetDate := time.Now().UTC().Add(24 * time.Hour)
	day := isoWeekday(targetDate)
	testEnv.createSchedule(t, adminToken, roomID, []int{day}, "13:00", "14:00")

	slotID := testEnv.getFirstAvailableSlot(t, userToken, roomID, targetDate.Format("2006-01-02"))
	bookingID, _ := testEnv.createBooking(t, userToken, slotID)

	status1 := testEnv.cancelBooking(t, userToken, bookingID)
	status2 := testEnv.cancelBooking(t, userToken, bookingID)

	if status1 != "cancelled" || status2 != "cancelled" {
		t.Fatalf("expected both cancel calls to return status=cancelled, got %q and %q", status1, status2)
	}
}

type e2eEnv struct {
	baseURL string
	server  *httptest.Server
	db      *pgxpool.Pool
	client  *http.Client
}

func newE2EEnv(t *testing.T) *e2eEnv {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = loadDatabaseURLFromEnvFile(t)
	}
	if dsn == "" {
		t.Skip("DATABASE_URL must be set for e2e tests")
	}

	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("cannot connect to db: %v", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		db.Close()
		t.Fatalf("database not available: %v", err)
	}

	if _, err := db.Exec(context.Background(), `SELECT pg_advisory_lock($1)`, sharedTestDBLockID); err != nil {
		db.Close()
		t.Fatalf("acquire db lock: %v", err)
	}

	applyMigrations(t, db)
	resetData(t, db)
	seedDummyUsers(t, db)

	validate := validator.New()
	tokenService := auth.NewJWTService()

	userRepo := postgres.NewUserRepository(db)
	roomRepo := postgres.NewRoomRepository(db)
	scheduleRepo := postgres.NewScheduleRepository(db)
	slotRepo := postgres.NewSlotRepository(db)
	bookingRepo := postgres.NewBookingRepository(db)

	userService := service.NewUserService(userRepo, validate, tokenService)
	roomService := service.NewRoomService(roomRepo)
	scheduleService := service.NewScheduleService(scheduleRepo)
	slotService := service.NewSlotService(slotRepo)
	bookingService := service.NewBookingService(bookingRepo)

	userHandler := handler.NewUserHandler(userService)
	roomHandler := handler.NewRoomHandler(roomService)
	scheduleHandler := handler.NewScheduleHandler(scheduleService, validate)
	slotHandler := handler.NewSlotHandler(*slotService)
	bookingHandler := handler.NewBookingHandler(bookingService, validate)

	h := server.NewServer(tokenService, userHandler, roomHandler, scheduleHandler, slotHandler, bookingHandler)
	ts := httptest.NewServer(h)

	return &e2eEnv{
		baseURL: ts.URL,
		server:  ts,
		db:      db,
		client:  ts.Client(),
	}
}

func loadDatabaseURLFromEnvFile(t *testing.T) string {
	t.Helper()

	root := projectRoot(t)
	content, err := os.ReadFile(filepath.Join(root, ".env"))
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "DATABASE_URL=") {
			return strings.TrimPrefix(line, "DATABASE_URL=")
		}
	}

	return ""
}

func (e *e2eEnv) close() {
	_, _ = e.db.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, sharedTestDBLockID)
	e.server.Close()
	e.db.Close()
}

func (e *e2eEnv) dummyLogin(t *testing.T, role string) string {
	payload := map[string]any{"role": role}
	resp := e.doJSON(t, http.MethodPost, "/dummyLogin", "", payload)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("dummyLogin expected 200, got %d", resp.StatusCode)
	}

	body := decodeBody(t, resp.Body)
	token, _ := body["token"].(string)
	if token == "" {
		t.Fatalf("dummyLogin token is empty")
	}
	return token
}

func (e *e2eEnv) createRoom(t *testing.T, adminToken string) string {
	payload := map[string]any{
		"name":        "E2E Room",
		"description": "test room",
		"capacity":    6,
	}
	resp := e.doJSON(t, http.MethodPost, "/rooms/create", adminToken, payload)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create room expected 201, got %d", resp.StatusCode)
	}

	body := decodeBody(t, resp.Body)
	room := toMap(body["room"])
	id := pickString(room, "id", "ID")
	if id == "" {
		t.Fatalf("create room: id is empty")
	}
	return id
}

func (e *e2eEnv) createSchedule(t *testing.T, adminToken, roomID string, days []int, start, end string) {
	payload := map[string]any{
		"daysOfWeek": days,
		"startTime":  start,
		"endTime":    end,
	}
	path := fmt.Sprintf("/rooms/%s/schedule/create", roomID)
	resp := e.doJSON(t, http.MethodPost, path, adminToken, payload)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create schedule expected 201, got %d", resp.StatusCode)
	}
}

func (e *e2eEnv) getFirstAvailableSlot(t *testing.T, token, roomID, date string) string {
	path := fmt.Sprintf("/rooms/%s/slots/list?date=%s", roomID, date)
	resp := e.doJSON(t, http.MethodGet, path, token, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get slots expected 200, got %d", resp.StatusCode)
	}

	body := decodeBody(t, resp.Body)
	rawSlots, ok := body["slots"].([]any)
	if !ok || len(rawSlots) == 0 {
		t.Fatalf("expected non-empty slots list")
	}

	slot := toMap(rawSlots[0])
	slotID := pickString(slot, "id", "ID")
	if slotID == "" {
		t.Fatalf("slot id is empty")
	}
	return slotID
}

func (e *e2eEnv) createBooking(t *testing.T, userToken, slotID string) (string, string) {
	payload := map[string]any{"slotId": slotID}
	resp := e.doJSON(t, http.MethodPost, "/bookings/create", userToken, payload)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create booking expected 201, got %d", resp.StatusCode)
	}

	body := decodeBody(t, resp.Body)
	booking := toMap(body["booking"])
	id := pickString(booking, "id", "ID")
	status := pickString(booking, "status", "Status")
	return id, status
}

func (e *e2eEnv) cancelBooking(t *testing.T, userToken, bookingID string) string {
	path := fmt.Sprintf("/bookings/%s/cancel", bookingID)
	resp := e.doJSON(t, http.MethodPost, path, userToken, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("cancel booking expected 200, got %d", resp.StatusCode)
	}

	body := decodeBody(t, resp.Body)
	booking := toMap(body["booking"])
	return pickString(booking, "status", "Status")
}

func (e *e2eEnv) doJSON(t *testing.T, method, path, token string, payload any) *http.Response {
	t.Helper()

	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		body = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, e.baseURL+path, body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		t.Fatalf("http call failed: %v", err)
	}
	return resp
}

func decodeBody(t *testing.T, r io.ReadCloser) map[string]any {
	t.Helper()
	defer r.Close()
	var m map[string]any
	if err := json.NewDecoder(r).Decode(&m); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return m
}

func toMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func pickString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

func isoWeekday(d time.Time) int {
	w := int(d.Weekday())
	if w == 0 {
		return 7
	}
	return w
}

func applyMigrations(t *testing.T, db *pgxpool.Pool) {
	t.Helper()
	root := projectRoot(t)
	files, err := filepath.Glob(filepath.Join(root, "migrations", "*.up.sql"))
	if err != nil {
		t.Fatalf("glob migrations: %v", err)
	}
	sort.Strings(files)
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read migration %s: %v", f, err)
		}
		if _, err := db.Exec(context.Background(), string(content)); err != nil {
			t.Fatalf("apply migration %s: %v", filepath.Base(f), err)
		}
	}
}

func resetData(t *testing.T, db *pgxpool.Pool) {
	t.Helper()
	_, err := db.Exec(context.Background(), `
		TRUNCATE TABLE bookings, slots, schedules, rooms RESTART IDENTITY CASCADE;
	`)
	if err != nil {
		t.Fatalf("reset data: %v", err)
	}
}

func seedDummyUsers(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	_, err := db.Exec(context.Background(), `
		INSERT INTO users (id, email, password_hash, role)
		VALUES
			($1, 'admin@example.com', 'dummy', 'admin'),
			($2, 'user@example.com', 'dummy', 'user')
		ON CONFLICT (id) DO NOTHING
	`, auth.AdminDummyUserID, auth.UserDummyUserID)
	if err != nil {
		t.Fatalf("seed dummy users: %v", err)
	}
}

func projectRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("cannot determine current file path")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
