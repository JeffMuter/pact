-- use this to override existing tables, keep in mind, data will poof
DROP TABLE IF EXISTS test;
DROP TABLE IF EXISTS support_tickets;
DROP TABLE IF EXISTS task_submissions;
DROP TABLE IF EXISTS assigned_tasks;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS connection_requests;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS connections;
DROP TABLE IF EXISTS users;

-- Users table
CREATE TABLE users (
    user_id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    active_connection_id INTEGER DEFAULT NULL,
    is_member INTEGER NOT NULL DEFAULT 0,
    stripe_customer_id TEXT UNIQUE DEFAULT NULL,
    stripe_subscription_id TEXT UNIQUE DEFAULT NULL,
    subscription_status TEXT DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (active_connection_id) REFERENCES connections(connection_id)
);

-- Connections represent an established manager-worker relationship.
-- A user can appear in multiple connections with different partners, and can
-- hold opposite roles in different connections (manager in one, worker in another).
-- worker_points tracks the worker's earned points within this connection.
CREATE TABLE connections (
    connection_id INTEGER PRIMARY KEY AUTOINCREMENT,
    manager_id INTEGER NOT NULL,
    worker_id INTEGER NOT NULL,
    worker_points INTEGER NOT NULL DEFAULT 0,
    UNIQUE(manager_id, worker_id),
    FOREIGN KEY (manager_id) REFERENCES users(user_id),
    FOREIGN KEY (worker_id) REFERENCES users(user_id)
);

CREATE TABLE connection_requests (
    request_id INTEGER PRIMARY KEY AUTOINCREMENT,
    sender_id INTEGER NOT NULL,
    receiver_id INTEGER NOT NULL,
    suggested_worker_id INTEGER NOT NULL,
    suggested_manager_id INTEGER NOT NULL,
    is_active INTEGER NOT NULL DEFAULT 1,
    FOREIGN KEY (sender_id) REFERENCES users(user_id),
    FOREIGN KEY (receiver_id) REFERENCES users(user_id),
    FOREIGN KEY (suggested_manager_id) REFERENCES users(user_id),
    FOREIGN KEY (suggested_worker_id) REFERENCES users(user_id),
    CHECK (sender_id != receiver_id),
    CHECK (suggested_manager_id != suggested_worker_id),
    CHECK (
        (sender_id IN (suggested_manager_id, suggested_worker_id)) AND
        (receiver_id IN (suggested_manager_id, suggested_worker_id))
    ),
    UNIQUE(sender_id, receiver_id)
);

-- Sessions table
CREATE TABLE sessions (
    session_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token TEXT NOT NULL,
    created_at TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Tasks are reusable templates owned by a manager (not tied to a connection).
-- type: 'normal', 'punishment', 'reward'
-- is_bookmarked: manager can save for later assignment
-- repeat_frequency: NULL = no repeat, 'daily', 'weekly', 'monthly'
-- point_cost: only used for reward-type tasks (what the worker pays to purchase)
-- punishment_task_id: optional default punishment template to auto-assign on disapproval
CREATE TABLE tasks (
    task_id INTEGER PRIMARY KEY AUTOINCREMENT,
    manager_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    type TEXT NOT NULL DEFAULT 'normal' CHECK (type IN ('normal', 'punishment', 'reward')),
    default_points INTEGER NOT NULL DEFAULT 20,
    default_duration_minutes INTEGER NOT NULL DEFAULT 1440,
    timer_days INTEGER DEFAULT NULL,
    timer_hours INTEGER DEFAULT NULL,
    timer_minutes INTEGER DEFAULT NULL,
    requires_image INTEGER NOT NULL DEFAULT 0,
    num_images_required INTEGER NOT NULL DEFAULT 1,
    requires_video INTEGER NOT NULL DEFAULT 0,
    num_videos_required INTEGER NOT NULL DEFAULT 1,
    requires_audio INTEGER NOT NULL DEFAULT 0,
    num_audio_required INTEGER NOT NULL DEFAULT 1,
    min_word_count INTEGER,
    point_cost INTEGER,
    is_bookmarked INTEGER NOT NULL DEFAULT 0,
    repeat_frequency TEXT CHECK (repeat_frequency IN ('daily', 'weekly', 'monthly') OR repeat_frequency IS NULL),
    repeat_connection_id INTEGER,
    last_assigned_at TIMESTAMP,
    punishment_task_id INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (manager_id) REFERENCES users(user_id),
    FOREIGN KEY (punishment_task_id) REFERENCES tasks(task_id),
    FOREIGN KEY (repeat_connection_id) REFERENCES connections(connection_id)
);

-- Assigned tasks are concrete instances of a task template, tied to a connection.
-- status: 'todo', 'in_review', 'completed', 'failed'
-- type mirrors the template type at assignment time for easier querying.
-- assigned_at + duration_minutes define the timer; due_time is the computed deadline.
-- punishment_task_id: the punishment template to auto-assign if this task is disapproved.
CREATE TABLE assigned_tasks (
    assigned_task_id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    connection_id INTEGER NOT NULL,
    worker_id INTEGER NOT NULL,
    type TEXT NOT NULL DEFAULT 'normal' CHECK (type IN ('normal', 'punishment', 'reward')),
    status TEXT NOT NULL DEFAULT 'todo' CHECK (status IN ('todo', 'in_review', 'completed', 'failed')),
    points INTEGER NOT NULL,
    duration_minutes INTEGER NOT NULL DEFAULT 1440,
    timer_days INTEGER DEFAULT NULL,
    timer_hours INTEGER DEFAULT NULL,
    timer_minutes INTEGER DEFAULT NULL,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    due_time TIMESTAMP NOT NULL,
    requires_image INTEGER NOT NULL DEFAULT 0,
    num_images_required INTEGER NOT NULL DEFAULT 1,
    requires_video INTEGER NOT NULL DEFAULT 0,
    num_videos_required INTEGER NOT NULL DEFAULT 1,
    requires_audio INTEGER NOT NULL DEFAULT 0,
    num_audio_required INTEGER NOT NULL DEFAULT 1,
    min_word_count INTEGER,
    punishment_task_id INTEGER,
    completed_at TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(task_id),
    FOREIGN KEY (connection_id) REFERENCES connections(connection_id),
    FOREIGN KEY (worker_id) REFERENCES users(user_id),
    FOREIGN KEY (punishment_task_id) REFERENCES tasks(task_id)
);

-- Worker submissions for assigned tasks.
-- A worker fills in text, and optionally attaches image/video/audio paths.
-- submitted_at is when the worker (or auto-submit timer) submitted the work.
-- image_paths, video_paths, audio_paths store JSON arrays of file paths
CREATE TABLE task_submissions (
    submission_id INTEGER PRIMARY KEY AUTOINCREMENT,
    assigned_task_id INTEGER NOT NULL UNIQUE,
    submission_text TEXT,
    image_paths TEXT,
    video_paths TEXT,
    audio_paths TEXT,
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (assigned_task_id) REFERENCES assigned_tasks(assigned_task_id)
);

-- Support tickets: User-submitted support requests
CREATE TABLE support_tickets (
    ticket_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    email TEXT NOT NULL,
    issue_description TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

-- Performance indexes for common query patterns
CREATE INDEX idx_assigned_tasks_connection_status ON assigned_tasks(connection_id, status);
CREATE INDEX idx_assigned_tasks_worker_status ON assigned_tasks(worker_id, status);
CREATE INDEX idx_connection_requests_receiver_active ON connection_requests(receiver_id, is_active);
CREATE INDEX idx_tasks_manager_type ON tasks(manager_id, type);
CREATE INDEX idx_tasks_repeat ON tasks(repeat_frequency, last_assigned_at) WHERE repeat_connection_id IS NOT NULL;
CREATE INDEX idx_support_tickets_user ON support_tickets(user_id);
