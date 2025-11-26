-- Add down migration script here
ALTER TABLE bookings DROP COLUMN is_notify_expired_sent;