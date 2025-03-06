-- Add is_deleted column to comments table
ALTER TABLE comments ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT FALSE;