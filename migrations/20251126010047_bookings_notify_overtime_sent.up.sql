-- Add up migration script here
ALTER TABLE bookings DROP COLUMN is_notify_expired_sent;

ALTER TABLE bookings
ADD COLUMN is_notify_overtime_sent BOOLEAN DEFAULT FALSE;