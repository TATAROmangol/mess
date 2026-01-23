CREATE TABLE last_read_outbox (
    id SERIAL PRIMARY KEY,
    recipient_id TEXT NOT NULL,
    chat_id INT NOT NULL,
    subject_id TEXT NOT NULL,
    message_id INT NOT NULL,
    deleted_at TIMESTAMPTZ
);