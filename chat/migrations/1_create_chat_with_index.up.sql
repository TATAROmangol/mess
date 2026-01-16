CREATE TABLE chat (
    id TEXT PRIMARY KEY,
    first_subject_id TEXT NOT NULL,
    second_subject_id TEXT NOT NULL,
    messages_count INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
)

CREATE INDEX idx_chat_subjects_not_deleted
ON chat (first_subject_id, second_subject_id)
WHERE deleted_at IS NULL;