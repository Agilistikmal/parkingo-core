-- Add up migration script here
ALTER TABLE bookings
ADD COLUMN is_notify_expired_sent BOOLEAN DEFAULT FALSE;