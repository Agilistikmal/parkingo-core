-- Add up migration script here
ALTER TABLE parkings ADD COLUMN slug VARCHAR(255) NOT NULL;