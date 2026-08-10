CREATE TABLE chern(
    user_id BIGINT PRIMARY KEY,
    service_id INT REFERENCES services(id) ON DELETE CASCADE
);