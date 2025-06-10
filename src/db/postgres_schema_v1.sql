CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS postgis_topology;

-- Cars table
-- user_id is a "soft" foreign key, as it references the user_id in
-- the Auth service "users" table, which is not managed by the backend service.
CREATE TABLE IF NOT EXISTS cars (
    plate TEXT PRIMARY KEY,
    brand TEXT,
    model TEXT,
    user_id TEXT NOT NULL
);

-- Zones table
CREATE TABLE IF NOT EXISTS zones (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    available BOOLEAN NOT NULL DEFAULT TRUE,
    geometry GEOMETRY(MULTIPOLYGON, 4326) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    price_offset REAL NOT NULL DEFAULT 0.0,
    price_lin REAL NOT NULL DEFAULT 1.0,
    price_exp REAL NOT NULL DEFAULT 0.0,
    CONSTRAINT valid_geometry CHECK (ST_IsValid(geometry)),
    CONSTRAINT non_empty_geometry CHECK (ST_NPoints(geometry) > 0)
);

-- Zone User Roles table
-- user_id is a "soft" foreign key to Auth service users table
-- assigned_by is a "soft" foreign key to Auth service users table
CREATE TABLE IF NOT EXISTS zone_user_roles (
    id SERIAL PRIMARY KEY,
    zone_id INTEGER NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'controller')),
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    assigned_by TEXT NOT NULL
);

-- Tickets table
CREATE TABLE IF NOT EXISTS tickets (
    id SERIAL PRIMARY KEY,
    zone_id INTEGER NOT NULL,
    plate TEXT NOT NULL,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    price REAL NOT NULL,
    paid BOOLEAN NOT NULL DEFAULT FALSE,
    creation_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (plate) REFERENCES cars(plate) ON DELETE CASCADE,
    FOREIGN KEY (zone_id) REFERENCES zones(id) ON DELETE CASCADE
);

-- Fines table
CREATE TABLE IF NOT EXISTS fines (
    id SERIAL PRIMARY KEY,
    zone_id INTEGER NOT NULL,
    plate TEXT NOT NULL,
    date TIMESTAMP NOT NULL,
    amount REAL NOT NULL,
    paid BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY (plate) REFERENCES cars(plate) ON DELETE CASCADE,
    FOREIGN KEY (zone_id) REFERENCES zones(id) ON DELETE CASCADE
);
