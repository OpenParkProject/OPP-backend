-- Cars table
-- user_id is a "soft" foreign key, as it references the user_id in
-- the Auth service "users" table, which is not managed by the backend service.
CREATE TABLE IF NOT EXISTS cars (
    plate TEXT PRIMARY KEY,
    brand TEXT,
    model TEXT,
    user_id TEXT NOT NULL
);

-- Tickets table
CREATE TABLE IF NOT EXISTS tickets (
    id SERIAL PRIMARY KEY,
    plate TEXT NOT NULL,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    price REAL NOT NULL,
    paid BOOLEAN NOT NULL DEFAULT FALSE,
    creation_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (plate) REFERENCES cars(plate) ON DELETE CASCADE
);

-- Fines table
CREATE TABLE IF NOT EXISTS fines (
    id SERIAL PRIMARY KEY,
    plate TEXT NOT NULL,
    date TIMESTAMP NOT NULL,
    amount REAL NOT NULL,
    paid BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY (plate) REFERENCES cars(plate) ON DELETE CASCADE
);

