-- Add down migration script here
ALTER TABLE parking_slots  
DROP COLUMN row,  
DROP COLUMN col;
