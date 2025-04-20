-- Add up migration script here
CREATE INDEX idx_payment_reference ON bookings (payment_reference);