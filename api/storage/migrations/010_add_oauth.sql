-- Migration: add_oauth_columns_to_users

-- Add OAuth provider columns to users table
ALTER TABLE users
    ADD COLUMN provider VARCHAR(50) NULL,
    ADD COLUMN provider_id VARCHAR(255) NULL;

-- Comment: provider stores the OAuth provider name (google, github, etc.)
-- provider_id stores the unique user ID from the provider

-- Create an index for efficient OAuth user lookups
CREATE INDEX idx_users_oauth ON users(provider, provider_id);

-- Create an index for email lookups during OAuth flow
CREATE INDEX idx_users_email ON users(email);
