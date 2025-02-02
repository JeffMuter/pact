-- use this to override existing tables, keep in mind, data will poof
DROP TABLE IF EXISTS test;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS connections;
DROP TABLE IF EXISTS connection_requests;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS rewards;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS punishments;
DROP TABLE IF EXISTS assigned_tasks;
DROP TABLE IF EXISTS assigned_punishments;
DROP TABLE IF EXISTS task_workermissions;

-- Users table
CREATE TABLE users (
    user_id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT CHECK(role IN ('manager', 'worker')) NOT NULL,
    is_member INTEGER NOT NULL DEFAULT 0,
    points INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE connections (
    connection_id INTEGER PRIMARY KEY AUTOINCREMENT,
    manager_id INTEGER NOT NULL,
    worker_id INTEGER NOT NULL,
    UNIQUE(manager_id, worker_id),
    FOREIGN KEY (manager_id) REFERENCES users(user_id),
    FOREIGN KEY (worker_id) REFERENCES users(user_id)
);

CREATE TABLE connection_requests (
    request_id INTEGER PRIMARY KEY AUTOINCREMENT,
    sender_id INTEGER NOT NULL,
    reciever_id INTEGER NOT NULL,
    suggested_worker_id INTEGER NOT NULL,
    suggested_manager_id INTEGER NOT NULL,
    is_active INTEGER NOT NULL DEFAULT 1,
    FOREIGN KEY (sender_id) REFERENCES users(user_id),
    FOREIGN KEY (reciever_id) REFERENCES users(user_id),
    FOREIGN KEY (suggested_manager_id) REFERENCES users(user_id),
    FOREIGN KEY (suggested_worker_id) REFERENCES users(user_id),
    CHECK (sender_id != reciever_id),
    CHECK (suggested_manager_id != suggested_worker_id),
    CHECK (
        (sender_id IN (suggested_manager_id, suggested_worker_id)) AND
        (reciever_id IN (suggested_manager_id, suggested_worker_id))
    ),
    UNIQUE(sender_id, reciever_id)
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


-- Rewards table
CREATE TABLE rewards (
    reward_id INTEGER PRIMARY KEY AUTOINCREMENT,
    manager_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    point_cost INTEGER NOT NULL,
    FOREIGN KEY (manager_id) REFERENCES users(user_id)
);

-- Tasks table
CREATE TABLE tasks (
    task_id INTEGER PRIMARY KEY AUTOINCREMENT,
    manager_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    default_points INTEGER NOT NULL DEFAULT 20,
    default_duration INTEGER NOT NULL DEFAULT 1440, -- 24 hours in minutes
    requires_image BOOLEAN NOT NULL DEFAULT 1,
    requires_video BOOLEAN NOT NULL DEFAULT 0,
    word_count INTEGER,
    FOREIGN KEY (manager_id) REFERENCES users(user_id)
);

-- Punishments table
CREATE TABLE punishments (
    punishment_id INTEGER PRIMARY KEY AUTOINCREMENT,
    manager_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    FOREIGN KEY (manager_id) REFERENCES users(user_id)
);

-- Assigned Tasks table
CREATE TABLE assigned_tasks (
    assigned_task_id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    worker_id INTEGER NOT NULL,
    points INTEGER NOT NULL,
    due_time TIMESTAMP NOT NULL,
    requires_image BOOLEAN NOT NULL,
    requires_video BOOLEAN NOT NULL,
    word_count INTEGER,
    completed_at TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(task_id),
    FOREIGN KEY (worker_id) REFERENCES users(user_id)
);

-- Assigned Punishments table
CREATE TABLE assigned_punishments (
    assigned_punishment_id INTEGER PRIMARY KEY AUTOINCREMENT,
    punishment_id INTEGER NOT NULL,
    worker_id INTEGER NOT NULL,
    assigned_task_id INTEGER NOT NULL,
    completed_at TIMESTAMP,
    FOREIGN KEY (punishment_id) REFERENCES punishments(punishment_id),
    FOREIGN KEY (worker_id) REFERENCES users(user_id),
    FOREIGN KEY (assigned_task_id) REFERENCES assigned_tasks(assigned_task_id)
);

-- Task workermissions table
CREATE TABLE task_workermissions (
    workermission_id INTEGER PRIMARY KEY AUTOINCREMENT,
    assigned_task_id INTEGER NOT NULL,
    workermission_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    image_path TEXT,
    video_path TEXT,
    text_content TEXT,
    FOREIGN KEY (assigned_task_id) REFERENCES assigned_tasks(assigned_task_id)
);
