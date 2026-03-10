-- Migration: Buckets feature schema redesign
-- This migration drops and recreates tables for the buckets feature.
-- WARNING: This will destroy all existing data in affected tables.
-- Back up database/database.db before running if you need to preserve data.
--
-- Run with: sqlite3 database/database.db < database/migrations/001_buckets_schema.sql

-- Drop tables in dependency order (children first)
DROP TABLE IF EXISTS test;
DROP TABLE IF EXISTS task_workermissions;
DROP TABLE IF EXISTS assigned_punishments;
DROP TABLE IF EXISTS assigned_tasks;
DROP TABLE IF EXISTS punishments;
DROP TABLE IF EXISTS rewards;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS task_submissions;

-- Drop the old points column from users (SQLite doesn't support DROP COLUMN
-- before 3.35.0, so we recreate the table)
-- First, preserve existing user data
CREATE TABLE IF NOT EXISTS users_backup AS SELECT
    user_id, email, username, password_hash, active_connection_id, is_member, created_at
FROM users;

DROP TABLE IF EXISTS users;

CREATE TABLE users (
    user_id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    active_connection_id INTEGER DEFAULT NULL,
    is_member INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (active_connection_id) REFERENCES connections(connection_id)
);

-- Restore user data (without points column)
INSERT INTO users (user_id, email, username, password_hash, active_connection_id, is_member, created_at)
SELECT user_id, email, username, password_hash, active_connection_id, is_member, created_at
FROM users_backup;

DROP TABLE IF EXISTS users_backup;

-- Recreate connections with worker_points column
CREATE TABLE IF NOT EXISTS connections_backup AS SELECT
    connection_id, manager_id, worker_id
FROM connections;

DROP TABLE IF EXISTS connections;

CREATE TABLE connections (
    connection_id INTEGER PRIMARY KEY AUTOINCREMENT,
    manager_id INTEGER NOT NULL,
    worker_id INTEGER NOT NULL,
    worker_points INTEGER NOT NULL DEFAULT 0,
    UNIQUE(manager_id, worker_id),
    FOREIGN KEY (manager_id) REFERENCES users(user_id),
    FOREIGN KEY (worker_id) REFERENCES users(user_id)
);

INSERT INTO connections (connection_id, manager_id, worker_id, worker_points)
SELECT connection_id, manager_id, worker_id, 0
FROM connections_backup;

DROP TABLE IF EXISTS connections_backup;

-- Create new tasks template table
CREATE TABLE tasks (
    task_id INTEGER PRIMARY KEY AUTOINCREMENT,
    manager_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    type TEXT NOT NULL DEFAULT 'normal' CHECK (type IN ('normal', 'punishment', 'reward')),
    default_points INTEGER NOT NULL DEFAULT 20,
    default_duration_minutes INTEGER NOT NULL DEFAULT 1440,
    requires_image INTEGER NOT NULL DEFAULT 0,
    requires_video INTEGER NOT NULL DEFAULT 0,
    requires_audio INTEGER NOT NULL DEFAULT 0,
    min_word_count INTEGER,
    point_cost INTEGER,
    is_bookmarked INTEGER NOT NULL DEFAULT 0,
    repeat_frequency TEXT CHECK (repeat_frequency IN ('daily', 'weekly', 'monthly') OR repeat_frequency IS NULL),
    punishment_task_id INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (manager_id) REFERENCES users(user_id),
    FOREIGN KEY (punishment_task_id) REFERENCES tasks(task_id)
);

-- Create new assigned_tasks table
CREATE TABLE assigned_tasks (
    assigned_task_id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    connection_id INTEGER NOT NULL,
    worker_id INTEGER NOT NULL,
    type TEXT NOT NULL DEFAULT 'normal' CHECK (type IN ('normal', 'punishment', 'reward')),
    status TEXT NOT NULL DEFAULT 'todo' CHECK (status IN ('todo', 'in_review', 'completed', 'failed')),
    points INTEGER NOT NULL,
    duration_minutes INTEGER NOT NULL DEFAULT 1440,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    due_time TIMESTAMP NOT NULL,
    requires_image INTEGER NOT NULL DEFAULT 0,
    requires_video INTEGER NOT NULL DEFAULT 0,
    requires_audio INTEGER NOT NULL DEFAULT 0,
    min_word_count INTEGER,
    punishment_task_id INTEGER,
    completed_at TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(task_id),
    FOREIGN KEY (connection_id) REFERENCES connections(connection_id),
    FOREIGN KEY (worker_id) REFERENCES users(user_id),
    FOREIGN KEY (punishment_task_id) REFERENCES tasks(task_id)
);

-- Create new task_submissions table
CREATE TABLE task_submissions (
    submission_id INTEGER PRIMARY KEY AUTOINCREMENT,
    assigned_task_id INTEGER NOT NULL UNIQUE,
    submission_text TEXT,
    image_path TEXT,
    video_path TEXT,
    audio_path TEXT,
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (assigned_task_id) REFERENCES assigned_tasks(assigned_task_id)
);
