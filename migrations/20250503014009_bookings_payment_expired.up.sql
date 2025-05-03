-- Add up migration script here
ALTER TABLE bookings ADD COLUMN payment_expired_at TIMESTAMP NULL DEFAULT NULL;