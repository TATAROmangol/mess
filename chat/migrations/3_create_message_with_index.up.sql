CREATE TABLE message (
    id TEXT PRIMARY KEY,
    chat_id TEXT NOT NULL,
    sender_subject_id TEXT NOT NULL,
    content TEXT NOT NULL,
    number INT NOT NULL,
    version int NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
)

CREATE INDEX idx_message_chatid
ON message (chat_id);

CREATE INDEX idx_message_number
ON message (number);