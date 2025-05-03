-- Add down migration script here
ALTER TABLE bookings DROP COLUMN payment_expired_at;