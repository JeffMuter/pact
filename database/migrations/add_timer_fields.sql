-- Migration: Add timer fields to tasks and assigned_tasks tables
-- Date: 2026-03-04
-- Description: Adds timer_days, timer_hours, timer_minutes to support granular time control

-- Add timer fields to tasks table
ALTER TABLE tasks ADD COLUMN timer_days INTEGER DEFAULT NULL;
ALTER TABLE tasks ADD COLUMN timer_hours INTEGER DEFAULT NULL;
ALTER TABLE tasks ADD COLUMN timer_minutes INTEGER DEFAULT NULL;

-- Add timer fields to assigned_tasks table
ALTER TABLE assigned_tasks ADD COLUMN timer_days INTEGER DEFAULT NULL;
ALTER TABLE assigned_tasks ADD COLUMN timer_hours INTEGER DEFAULT NULL;
ALTER TABLE assigned_tasks ADD COLUMN timer_minutes INTEGER DEFAULT NULL;

-- Optional: Migrate existing duration_minutes to timer fields
-- This converts existing durations to human-readable timer format
-- Uncomment the following lines if you want to migrate existing data:

-- UPDATE tasks
-- SET 
--     timer_days = duration_minutes / 1440,
--     timer_hours = (duration_minutes % 1440) / 60,
--     timer_minutes = duration_minutes % 60
-- WHERE duration_minutes IS NOT NULL AND timer_days IS NULL;

-- UPDATE assigned_tasks
-- SET 
--     timer_days = duration_minutes / 1440,
--     timer_hours = (duration_minutes % 1440) / 60,
--     timer_minutes = duration_minutes % 60
-- WHERE duration_minutes IS NOT NULL AND timer_days IS NULL;
