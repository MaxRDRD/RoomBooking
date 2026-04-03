package unit_test

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/model"
	"RoomBookingService/internal/usecase/service"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type roomRepoStub struct {
	createErr error
	all       []*model.Room
	allErr    error
	received  *model.Room
}

func (r *roomRepoStub) CreateRoom(_ context.Context, room *model.Room) error {
	r.received = room
	if r.received.ID == uuid.Nil {
		r.received.ID = uuid.New()
	}
	return r.createErr
}

func (r *roomRepoStub) GetAllRooms(_ context.Context) ([]*model.Room, error) {
	return r.all, r.allErr
}

func TestRoomService_CreateRoom_SetsFields(t *testing.T) {
	cap := 8
	repo := &roomRepoStub{}
	svc := service.NewRoomService(repo)

	res, err := svc.CreateRoom(context.Background(), dto.CreateRoomRequest{
		Name:        "A101",
		Description: "small room",
		Capacity:    &cap,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res == nil {
		t.Fatalf("expected result, got nil")
	}
	if repo.received == nil {
		t.Fatalf("expected repository to receive room")
	}
	if repo.received.Description != "small room" {
		t.Fatalf("expected description to be passed, got %q", repo.received.Description)
	}
	if repo.received.Capacity != 8 {
		t.Fatalf("expected capacity=8, got %d", repo.received.Capacity)
	}
}

func TestRoomService_CreateRoom_Error(t *testing.T) {
	repo := &roomRepoStub{createErr: errors.New("db error")}
	svc := service.NewRoomService(repo)

	_, err := svc.CreateRoom(context.Background(), dto.CreateRoomRequest{Name: "A101"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestRoomService_GetAllRooms(t *testing.T) {
	expected := []*model.Room{{ID: uuid.New(), Name: "A101"}}
	repo := &roomRepoStub{all: expected}
	svc := service.NewRoomService(repo)

	rooms, err := svc.GetAllRooms(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(rooms) != 1 {
		t.Fatalf("expected 1 room, got %d", len(rooms))
	}
}
