ALTER TABLE bookings
ALTER COLUMN booking_time TYPE timestamp without time zone
USING (date + booking_time)::timestamp;