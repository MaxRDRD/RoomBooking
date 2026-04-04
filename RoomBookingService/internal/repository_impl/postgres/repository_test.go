package postgres

import (
	"RoomBookingService/internal/auth"
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/model"
	myerrors "RoomBookingService/pkg/my_errors"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const sharedTestDBLockID int64 = 482913741

func TestRepositories_RoomScheduleSlotBookingFlow(t *testing.T) {
	db := newRepoTestDB(t)
	resetRepoData(t, db)
	seedRepoUsers(t, db)

	roomRepo := NewRoomRepository(db)
	scheduleRepo := NewScheduleRepository(db)
	slotRepo := NewSlotRepository(db)
	bookingRepo := NewBookingRepository(db)

	room := &model.Room{Name: "A101", Description: "main room", Capacity: 6}
	if err := roomRepo.CreateRoom(context.Background(), room); err != nil {
		t.Fatalf("create room: %v", err)
	}
	if room.ID == uuid.Nil {
		t.Fatal("expected room id")
	}

	rooms, err := roomRepo.GetAllRooms(context.Background())
	if err != nil {
		t.Fatalf("get all rooms: %v", err)
	}
	if len(rooms) != 1 || rooms[0].Description != "main room" {
		t.Fatalf("unexpected rooms result: %+v", rooms)
	}

	schedule := &model.Schedule{
		RoomID:     room.ID,
		DaysOfWeek: []int{isoWeekday(time.Now().UTC().Add(24 * time.Hour))},
		StartTime:  time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:    time.Date(0, 1, 1, 11, 0, 0, 0, time.UTC),
	}
	if err := scheduleRepo.CreateSchedule(context.Background(), schedule); err != nil {
		t.Fatalf("create schedule: %v", err)
	}
	if schedule.ID == uuid.Nil {
		t.Fatal("expected schedule id")
	}

	targetDate := time.Now().UTC().Add(24 * time.Hour)
	slots, err := slotRepo.GetAvailableSlots(context.Background(), dto.SlotsRequest{RoomID: room.ID, Date: targetDate})
	if err != nil {
		t.Fatalf("get available slots: %v", err)
	}
	if len(slots) == 0 {
		t.Fatal("expected generated slots")
	}

	booking, err := bookingRepo.CreateBooking(context.Background(), uuid.MustParse(auth.UserDummyUserID), dto.BookingRequest{SlotID: slots[0].ID})
	if err != nil {
		t.Fatalf("create booking: %v", err)
	}
	if booking.Status != "active" {
		t.Fatalf("unexpected booking status: %s", booking.Status)
	}

	_, err = bookingRepo.CreateBooking(context.Background(), uuid.MustParse(auth.UserDummyUserID), dto.BookingRequest{SlotID: slots[0].ID})
	if !errorsIsSlotBooked(err) {
		t.Fatalf("expected slot already booked error, got %v", err)
	}

	myBookings, err := bookingRepo.GetMyBookings(context.Background(), uuid.MustParse(auth.UserDummyUserID))
	if err != nil {
		t.Fatalf("get my bookings: %v", err)
	}
	if len(myBookings) != 1 {
		t.Fatalf("expected 1 booking, got %d", len(myBookings))
	}

	cancelled, err := bookingRepo.CancelBooking(context.Background(), uuid.MustParse(auth.UserDummyUserID), dto.CancelBookingRequest{BookingID: booking.ID})
	if err != nil {
		t.Fatalf("cancel booking: %v", err)
	}
	if cancelled.Status != "cancelled" {
		t.Fatalf("expected cancelled status, got %s", cancelled.Status)
	}

	cancelledAgain, err := bookingRepo.CancelBooking(context.Background(), uuid.MustParse(auth.UserDummyUserID), dto.CancelBookingRequest{BookingID: booking.ID})
	if err != nil {
		t.Fatalf("second cancel booking: %v", err)
	}
	if cancelledAgain.Status != "cancelled" {
		t.Fatalf("expected idempotent cancel, got %s", cancelledAgain.Status)
	}

	allBookings, total, err := bookingRepo.GetAllBookings(context.Background(), dto.BookingFilter{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("get all bookings: %v", err)
	}
	if total != 1 || len(allBookings) != 1 {
		t.Fatalf("unexpected admin booking list: total=%d len=%d", total, len(allBookings))
	}

	remaining, err := slotRepo.GetAvailableSlots(context.Background(), dto.SlotsRequest{RoomID: room.ID, Date: targetDate})
	if err != nil {
		t.Fatalf("get available slots after booking: %v", err)
	}
	if len(remaining) == 0 {
		t.Fatal("expected available slots to remain")
	}
}

func TestSlotRepository_NoScheduleAndWrongDay(t *testing.T) {
	db := newRepoTestDB(t)
	resetRepoData(t, db)
	seedRepoUsers(t, db)

	roomRepo := NewRoomRepository(db)
	slotRepo := NewSlotRepository(db)

	room := &model.Room{Name: "A102"}
	if err := roomRepo.CreateRoom(context.Background(), room); err != nil {
		t.Fatalf("create room: %v", err)
	}

	noSchedule, err := slotRepo.GetAvailableSlots(context.Background(), dto.SlotsRequest{RoomID: room.ID, Date: time.Now().UTC().Add(24 * time.Hour)})
	if err != nil {
		t.Fatalf("get slots without schedule: %v", err)
	}
	if len(noSchedule) != 0 {
		t.Fatalf("expected empty slots, got %d", len(noSchedule))
	}
}

func newRepoTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = loadRepoDSNFromEnvFile(t)
	}
	if dsn == "" {
		t.Skip("DATABASE_URL must be set for repository tests")
	}

	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	if err := db.Ping(context.Background()); err != nil {
		db.Close()
		t.Fatalf("ping db: %v", err)
	}

	if _, err := db.Exec(context.Background(), `SELECT pg_advisory_lock($1)`, sharedTestDBLockID); err != nil {
		db.Close()
		t.Fatalf("acquire db lock: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, sharedTestDBLockID)
		db.Close()
	})

	applyRepoMigrations(t, db)
	return db
}

func applyRepoMigrations(t *testing.T, db *pgxpool.Pool) {
	t.Helper()
	root := repoProjectRoot(t)
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

func resetRepoData(t *testing.T, db *pgxpool.Pool) {
	t.Helper()
	_, err := db.Exec(context.Background(), `TRUNCATE TABLE bookings, slots, schedules, rooms, users RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("reset data: %v", err)
	}
}

func seedRepoUsers(t *testing.T, db *pgxpool.Pool) {
	t.Helper()
	_, err := db.Exec(context.Background(), `
		INSERT INTO users (id, email, password_hash, role)
		VALUES
			($1, 'admin@example.com', 'dummy', 'admin'),
			($2, 'user@example.com', 'dummy', 'user')
		ON CONFLICT (id) DO NOTHING
	`, auth.AdminDummyUserID, auth.UserDummyUserID)
	if err != nil {
		t.Fatalf("seed users: %v", err)
	}
}

func loadRepoDSNFromEnvFile(t *testing.T) string {
	t.Helper()
	root := repoProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root, ".env"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "DATABASE_URL=") {
			return strings.TrimPrefix(line, "DATABASE_URL=")
		}
	}
	return ""
}

func repoProjectRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve current file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func isoWeekday(d time.Time) int {
	w := int(d.Weekday())
	if w == 0 {
		return 7
	}
	return w
}

func errorsIsSlotBooked(err error) bool {
	return err == myerrors.ErrSlotAlreadyBooked
}
