SHELL := /bin/sh

.PHONY: up seed down test

up:
	docker compose up -d --build

seed:
	docker compose exec -T postgres psql -U postgres -d room_booking -v ON_ERROR_STOP=1 -c "INSERT INTO users (id, email, password_hash, role, created_at) VALUES ('11111111-1111-1111-1111-111111111111', 'admin@example.com', 'dummy', 'admin', NOW()), ('22222222-2222-2222-2222-222222222222', 'user@example.com', 'dummy', 'user', NOW()) ON CONFLICT (id) DO NOTHING;"

down:
	docker compose down -v

test:
	cd RoomBookingService && go test ./... -cover