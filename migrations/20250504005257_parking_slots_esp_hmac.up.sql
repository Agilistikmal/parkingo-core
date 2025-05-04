-- Add up migration script here
ALTER TABLE parking_slots
ADD COLUMN esp_hmac VARCHAR(255) DEFAULT NULL,
ADD COLUMN preview_url VARCHAR(255) DEFAULT NULL;