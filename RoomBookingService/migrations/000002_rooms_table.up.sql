CREATE TABLE IF NOT EXISTS rooms(
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    capacity INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);