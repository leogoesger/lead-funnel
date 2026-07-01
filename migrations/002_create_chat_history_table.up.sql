-- Create chat_history table for conversation tracking
CREATE TABLE chat_history (
    id BIGSERIAL PRIMARY KEY,
    contact_id BIGINT NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    role VARCHAR(20) NOT NULL,
    -- Possible roles: user, assistant
    message_type VARCHAR(50),
    -- Possible types: text, question, answer, follow_up, tour_confirmation, etc.
    tokens_used INT,
    -- Track token usage for cost monitoring
    metadata JSONB,
    -- Store additional context like intent, sentiment, etc.
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient querying
CREATE INDEX idx_chat_history_contact_id ON chat_history(contact_id);
CREATE INDEX idx_chat_history_created_at ON chat_history(created_at DESC);
CREATE INDEX idx_chat_history_contact_created ON chat_history(contact_id, created_at DESC);
CREATE INDEX idx_chat_history_role ON chat_history(role);
