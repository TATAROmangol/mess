CREATE TABLE lastread (
    subject_id TEXT NOT NULL,
    chat_id TEXT NOT NULL,
    message_number INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
)

CREATE INDEX idx_lastread_subject_chat_not_deleted
ON lastread (subject_id, chat_id)
WHERE deleted_at IS NULL;