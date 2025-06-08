-- Add up migration script here
-- Add available_earnings, withdrawn_earnings, total_earnings, total_bookings column to parkings table
ALTER TABLE parkings
ADD COLUMN available_earnings DECIMAL(10, 2) NOT NULL DEFAULT 0;

ALTER TABLE parkings
ADD COLUMN withdrawn_earnings DECIMAL(10, 2) NOT NULL DEFAULT 0;

ALTER TABLE parkings
ADD COLUMN total_earnings DECIMAL(10, 2) NOT NULL DEFAULT 0;

ALTER TABLE parkings
ADD COLUMN total_bookings INT NOT NULL DEFAULT 0;