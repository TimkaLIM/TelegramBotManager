ALTER TABLE bookings
ALTER COLUMN booking_time TYPE time USING booking_time::time;