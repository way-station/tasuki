-- Runs once, when the postgres volume is first created.
-- The test database is separate so integration tests can reset it freely.
CREATE DATABASE tasuki_test OWNER app;
