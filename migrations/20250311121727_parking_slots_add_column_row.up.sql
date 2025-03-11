-- Add up migration script here
ALTER TABLE parking_slots 
ADD COLUMN row INT NOT NULL DEFAULT 0,
ADD COLUMN col INT NOT NULL DEFAULT 0;