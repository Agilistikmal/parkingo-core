-- Add up migration script here
CREATE TABLE bookings (
  id SERIAL PRIMARY KEY,
  user_id INT NOT NULL,
  parking_id INT NOT NULL,
  slot_id INT NOT NULL,
  plate_number VARCHAR(255) NOT NULL,
  start_at TIMESTAMP NOT NULL,
  end_at TIMESTAMP NOT NULL,
  total_hours INT NOT NULL,
  total_fee DECIMAL(10, 2) NOT NULL,
  status VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users (id),
  FOREIGN KEY (parking_id) REFERENCES parkings (id),
  FOREIGN KEY (slot_id) REFERENCES parking_slots (id)
);