-- Drop chat_history table and indexes
DROP INDEX IF EXISTS idx_chat_history_contact_created;
DROP INDEX IF EXISTS idx_chat_history_role;
DROP INDEX IF EXISTS idx_chat_history_created_at;
DROP INDEX IF EXISTS idx_chat_history_contact_id;
DROP TABLE IF EXISTS chat_history;
