-- Add down migration script here
ALTER TABLE parkings DROP COLUMN total_earnings;

ALTER TABLE parkings DROP COLUMN total_bookings;

ALTER TABLE parkings DROP COLUMN available_earnings;

ALTER TABLE parkings DROP COLUMN withdrawn_earnings;