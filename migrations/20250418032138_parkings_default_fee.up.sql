-- Add up migration script here
ALTER TABLE parkings ADD COLUMN default_fee DECIMAL(10,2) NOT NULL DEFAULT 10000;