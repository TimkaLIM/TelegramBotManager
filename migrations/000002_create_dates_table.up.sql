CREATE TABLE schedules (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL, 
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    slot_duration_min INT DEFAULT 60
);