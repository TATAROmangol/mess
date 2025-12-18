CREATE TABLE profile (
    subject_id TEXT PRIMARY KEY, 
    alias TEXT NOT NULL,
    avatar_url TEXT,
    version INT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
)
