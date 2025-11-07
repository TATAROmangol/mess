-- Active: 1760868796922@@localhost@5430@postgres@public
CREATE TABLE subject (
    id SERIAL PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP
);
