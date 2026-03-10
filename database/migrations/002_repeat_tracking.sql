-- Migration: Add repeat tracking columns to tasks table
-- Run with: sqlite3 database/database.db < database/migrations/002_repeat_tracking.sql

ALTER TABLE tasks ADD COLUMN repeat_connection_id INTEGER REFERENCES connections(connection_id);
ALTER TABLE tasks ADD COLUMN last_assigned_at TIMESTAMP;
