CREATE TABLE IF NOT EXISTS schedules(
    id SERIAL PRIMARY KEY,
    room_id INT NOT NULL UNIQUE,
    days_of_week INT[] NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

     CONSTRAINT fk_schedules_rooms
        FOREIGN KEY (room_id)
        REFERENCES rooms(id)
        ON DELETE CASCADE
);