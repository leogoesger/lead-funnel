-- Initialize database tables on first startup
-- This is run only if the database is empty

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Set timezone to UTC
ALTER DATABASE lead_funnel SET TIMEZONE='UTC';


-- Create contacts table for lead generation
CREATE TABLE contacts (
    id BIGSERIAL PRIMARY KEY,
    phone_number VARCHAR(20) NOT NULL UNIQUE,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    email VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'new',
    -- Possible statuses: new, contacted, tour_scheduled, enrolled, not_interested, inactive
    source VARCHAR(100),
    -- Where the lead came from (SMS, website, referral, etc.)
    notes TEXT,
    last_message_at TIMESTAMP,
    tour_scheduled_at TIMESTAMP,
    enrolled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for common queries
CREATE INDEX idx_contacts_phone_number ON contacts(phone_number);
CREATE INDEX idx_contacts_status ON contacts(status);
CREATE INDEX idx_contacts_created_at ON contacts(created_at DESC);
CREATE INDEX idx_contacts_last_message_at ON contacts(last_message_at DESC);
