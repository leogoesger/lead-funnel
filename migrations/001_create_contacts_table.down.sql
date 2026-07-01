-- Drop contacts table and indexes
DROP INDEX IF EXISTS idx_contacts_last_message_at;
DROP INDEX IF EXISTS idx_contacts_created_at;
DROP INDEX IF EXISTS idx_contacts_status;
DROP INDEX IF EXISTS idx_contacts_phone_number;
DROP TABLE IF EXISTS contacts;
