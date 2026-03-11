-- name: GetAllUsers :many
SELECT * FROM users;

-- name: GetUserById :one
SELECT * from users WHERE user_id = ?;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ?;

-- name: GetUsernameByUserId :one
SELECT username FROM users WHERE user_id = ?;

-- name: UserIsMemberById :one
SELECT is_member FROM users WHERE user_id = ?;

-- name: CreateUser :one
INSERT INTO users (email, username, password_hash) VALUES (?, ?, ?) returning user_id;

-- name: CreateSession :exec
INSERT INTO sessions(user_id, token, created_at, expires_at) VALUES(?, ?, ?, ?);

-- name: CreateRequest :exec
INSERT INTO connection_requests (sender_id, receiver_id, suggested_manager_id, suggested_worker_id) VALUES (?, ?, ?, ?);

-- name: GetUserPendingRequests :many
SELECT connection_requests.request_id, users.email, connection_requests.sender_id, connection_requests.receiver_id, connection_requests.suggested_manager_id, connection_requests.suggested_worker_id
FROM connection_requests
JOIN users ON connection_requests.sender_id = users.user_id
WHERE connection_requests.is_active = 1
AND connection_requests.receiver_id = ?;

-- name: GetConnectionRequestById :one
SELECT request_id, sender_id, receiver_id, suggested_manager_id, suggested_worker_id, is_active
FROM connection_requests
WHERE request_id = ?;

-- name: DeactivateConnectionRequest :exec
UPDATE connection_requests SET is_active = 0 WHERE request_id = ?;

-- name: DeleteConnectionRequestByUserIds :exec
DELETE FROM connection_requests
WHERE (sender_id = ? AND receiver_id = ?)
   OR (sender_id = ? AND receiver_id = ?);

-- name: CreateConnection :one
INSERT INTO connections (manager_id, worker_id) VALUES (?, ?) RETURNING connection_id;

-- name: GetConnectionsById :many
SELECT 
    c.connection_id,
    c.manager_id,
    m.username AS manager_username,
    c.worker_id,
    w.username AS worker_username
FROM connections c
JOIN users m ON c.manager_id = m.user_id
JOIN users w ON c.worker_id = w.user_id
WHERE ? IN (c.manager_id, c.worker_id);

-- name: UpdateActiveConnection :exec
UPDATE users SET active_connection_id = ? WHERE user_id = ?;

-- name: GetActiveConnectionId :one
SELECT active_connection_id
FROM users
WHERE user_id= ?;

-- name: GetActiveConnectionDetails :one
SELECT worker_id, manager_id
FROM connections
WHERE connection_id = ?;

-- name: GetActiveConnectionUserDetails :one
SELECT 
  CASE WHEN c.manager_id = ? THEN c.worker_id ELSE c.manager_id END AS user_id,
  u.username,
  CASE WHEN c.manager_id = ? THEN 'worker' ELSE 'manager' END AS role
FROM connections c
JOIN users u ON u.user_id = CASE WHEN c.manager_id = ? THEN c.worker_id ELSE c.manager_id END
WHERE c.connection_id = (SELECT active_connection_id FROM users WHERE users.user_id = ?);

-- name: DeleteConnection :exec
DELETE FROM connections WHERE connection_id = ?;

-- name: ClearActiveConnectionIfMatch :exec
UPDATE users 
SET active_connection_id = NULL 
WHERE user_id = ? AND active_connection_id = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE user_id = ?;

-- name: GetAccountPageData :one
SELECT 
    u.email,
    u.username,
    u.created_at,
    u.is_member,
    (SELECT COUNT(*) FROM connections WHERE ?1 IN (manager_id, worker_id)) AS connection_count,
    (SELECT COUNT(*) FROM connection_requests WHERE receiver_id = ?1 AND is_active = 1) AS pending_request_count
FROM users u
WHERE u.user_id = ?1;

-- =====================
-- POINTS
-- =====================

-- name: GetUserRoleInConnection :one
SELECT 
  CASE WHEN manager_id = ? THEN 'manager' ELSE 'worker' END AS role
FROM connections
WHERE connection_id = ?;

-- name: GetWorkerPoints :one
SELECT worker_points FROM connections WHERE connection_id = ?;

-- name: AddWorkerPoints :exec
UPDATE connections SET worker_points = worker_points + ? WHERE connection_id = ?;

-- name: DeductWorkerPoints :exec
UPDATE connections SET worker_points = worker_points - ? WHERE connection_id = ?;

-- =====================
-- TASK TEMPLATES
-- =====================

-- name: CreateTask :one
INSERT INTO tasks (
    manager_id, title, description, type, default_points,
    default_duration_minutes, timer_days, timer_hours, timer_minutes,
    requires_image, num_images_required, requires_video, num_videos_required,
    requires_audio, num_audio_required, min_word_count, point_cost, is_bookmarked,
    repeat_frequency, punishment_task_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING task_id;

-- name: GetTaskById :one
SELECT * FROM tasks WHERE task_id = ?;

-- name: GetManagerTasks :many
SELECT * FROM tasks
WHERE manager_id = ?
ORDER BY created_at DESC;

-- name: GetManagerTasksByType :many
SELECT * FROM tasks
WHERE manager_id = ? AND type = ?
ORDER BY created_at DESC;

-- name: GetBookmarkedTasks :many
SELECT * FROM tasks
WHERE manager_id = ? AND is_bookmarked = 1
ORDER BY created_at DESC;

-- name: GetRepeatingTasks :many
SELECT * FROM tasks
WHERE manager_id = ? AND repeat_frequency IS NOT NULL
ORDER BY created_at DESC;

-- name: UpdateTask :exec
UPDATE tasks SET
    title = ?, description = ?, type = ?, default_points = ?,
    default_duration_minutes = ?, timer_days = ?, timer_hours = ?, timer_minutes = ?,
    requires_image = ?, num_images_required = ?, requires_video = ?, num_videos_required = ?,
    requires_audio = ?, num_audio_required = ?, min_word_count = ?, point_cost = ?,
    is_bookmarked = ?, repeat_frequency = ?, punishment_task_id = ?
WHERE task_id = ? AND manager_id = ?;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE task_id = ? AND manager_id = ?;

-- name: GetRewardTasks :many
SELECT * FROM tasks
WHERE manager_id = ? AND type = 'reward'
ORDER BY created_at DESC;

-- =====================
-- ASSIGNED TASKS
-- =====================

-- name: AssignTask :one
INSERT INTO assigned_tasks (
    task_id, connection_id, worker_id, type, points,
    duration_minutes, timer_days, timer_hours, timer_minutes, due_time,
    requires_image, num_images_required, requires_video, num_videos_required,
    requires_audio, num_audio_required, min_word_count, punishment_task_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING assigned_task_id;

-- name: GetAssignedTaskById :one
SELECT * FROM assigned_tasks WHERE assigned_task_id = ?;

-- name: GetAssignedTasksByConnectionAndStatus :many
SELECT
    at.assigned_task_id,
    at.task_id,
    at.connection_id,
    at.worker_id,
    at.type,
    at.status,
    at.points,
    at.duration_minutes,
    at.timer_days,
    at.timer_hours,
    at.timer_minutes,
    at.assigned_at,
    at.due_time,
    at.requires_image,
    at.num_images_required,
    at.requires_video,
    at.num_videos_required,
    at.requires_audio,
    at.num_audio_required,
    at.min_word_count,
    at.punishment_task_id,
    at.completed_at,
    t.title,
    t.description,
    ts.image_paths,
    ts.video_paths,
    ts.audio_paths,
    ts.submission_text
FROM assigned_tasks at
JOIN tasks t ON at.task_id = t.task_id
LEFT JOIN task_submissions ts ON at.assigned_task_id = ts.assigned_task_id
WHERE at.connection_id = ? AND at.status = ?
ORDER BY at.due_time ASC;

-- name: GetAssignedTasksForWorkerByStatus :many
SELECT
    at.assigned_task_id,
    at.task_id,
    at.connection_id,
    at.worker_id,
    at.type,
    at.status,
    at.points,
    at.duration_minutes,
    at.timer_days,
    at.timer_hours,
    at.timer_minutes,
    at.assigned_at,
    at.due_time,
    at.requires_image,
    at.num_images_required,
    at.requires_video,
    at.num_videos_required,
    at.requires_audio,
    at.num_audio_required,
    at.min_word_count,
    at.punishment_task_id,
    at.completed_at,
    t.title,
    t.description
FROM assigned_tasks at
JOIN tasks t ON at.task_id = t.task_id
WHERE at.worker_id = ? AND at.connection_id = ? AND at.status = ?
ORDER BY at.due_time ASC;

-- name: UpdateAssignedTaskStatus :exec
UPDATE assigned_tasks SET status = ? WHERE assigned_task_id = ?;

-- name: CompleteAssignedTask :exec
UPDATE assigned_tasks SET status = 'completed', completed_at = CURRENT_TIMESTAMP
WHERE assigned_task_id = ?;

-- name: FailAssignedTask :exec
UPDATE assigned_tasks SET status = 'failed', completed_at = CURRENT_TIMESTAMP
WHERE assigned_task_id = ?;

-- name: SubmitAssignedTask :exec
UPDATE assigned_tasks SET status = 'in_review'
WHERE assigned_task_id = ?;

-- name: GetExpiredTodoTasks :many
SELECT * FROM assigned_tasks
WHERE status = 'todo' AND due_time <= CURRENT_TIMESTAMP;

-- name: CountAssignedTasksByConnectionAndStatus :one
SELECT COUNT(*) FROM assigned_tasks
WHERE connection_id = ? AND status = ?;

-- =====================
-- TASK SUBMISSIONS
-- =====================

-- name: CreateSubmission :one
INSERT INTO task_submissions (assigned_task_id, submission_text, image_paths, video_paths, audio_paths)
VALUES (?, ?, ?, ?, ?)
RETURNING submission_id;

-- name: GetSubmissionByAssignedTaskId :one
SELECT * FROM task_submissions WHERE assigned_task_id = ?;

-- name: UpdateSubmission :exec
UPDATE task_submissions SET
    submission_text = ?, image_paths = ?, video_paths = ?, audio_paths = ?
WHERE assigned_task_id = ?;

-- name: UpsertSubmission :one
INSERT INTO task_submissions (assigned_task_id, submission_text, image_paths, video_paths, audio_paths)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(assigned_task_id) DO UPDATE SET
    submission_text = excluded.submission_text,
    image_paths = excluded.image_paths,
    video_paths = excluded.video_paths,
    audio_paths = excluded.audio_paths,
    submitted_at = CURRENT_TIMESTAMP
RETURNING submission_id;

-- name: GetSubmissionWithTask :one
SELECT
    ts.submission_id,
    ts.assigned_task_id,
    ts.submission_text,
    ts.image_paths,
    ts.video_paths,
    ts.audio_paths,
    ts.submitted_at,
    at.status,
    at.type,
    at.points,
    at.requires_image,
    at.num_images_required,
    at.requires_video,
    at.num_videos_required,
    at.requires_audio,
    at.num_audio_required,
    at.min_word_count,
    t.title,
    t.description
FROM task_submissions ts
JOIN assigned_tasks at ON ts.assigned_task_id = at.assigned_task_id
JOIN tasks t ON at.task_id = t.task_id
WHERE ts.assigned_task_id = ?;

-- name: GetSubmissionForManager :one
SELECT
    ts.submission_id,
    ts.assigned_task_id,
    ts.submission_text,
    ts.image_paths,
    ts.video_paths,
    ts.audio_paths,
    ts.submitted_at,
    at.status,
    at.type,
    at.points,
    at.connection_id,
    t.title,
    t.description,
    c.manager_id,
    c.worker_id
FROM task_submissions ts
JOIN assigned_tasks at ON ts.assigned_task_id = at.assigned_task_id
JOIN tasks t ON at.task_id = t.task_id
JOIN connections c ON at.connection_id = c.connection_id
WHERE ts.assigned_task_id = ?;

-- =====================
-- REWARDS (worker-facing)
-- =====================

-- name: GetAvailableRewards :many
SELECT t.task_id, t.title, t.description, t.point_cost,
       t.requires_image, t.num_images_required, t.requires_video, t.num_videos_required, t.requires_audio, t.num_audio_required,
       t.min_word_count, t.default_duration_minutes, t.timer_days, t.timer_hours, t.timer_minutes, t.default_points
FROM tasks t
JOIN connections c ON t.manager_id = c.manager_id
WHERE c.connection_id = ? AND t.type = 'reward'
ORDER BY t.point_cost ASC;

-- name: GetConnectionForBuckets :one
SELECT
    c.connection_id,
    c.manager_id,
    c.worker_id,
    c.worker_points,
    m.username AS manager_username,
    w.username AS worker_username
FROM connections c
JOIN users m ON c.manager_id = m.user_id
JOIN users w ON c.worker_id = w.user_id
WHERE c.connection_id = ?;

-- =====================
-- BOOKMARKS & REPEATS
-- =====================

-- name: GetAllDueRepeatingTasks :many
SELECT * FROM tasks
WHERE repeat_frequency IS NOT NULL
  AND repeat_connection_id IS NOT NULL
  AND (
    (repeat_frequency = 'daily' AND (last_assigned_at IS NULL OR last_assigned_at < datetime('now', '-1 day')))
    OR (repeat_frequency = 'weekly' AND (last_assigned_at IS NULL OR last_assigned_at < datetime('now', '-7 days')))
    OR (repeat_frequency = 'monthly' AND (last_assigned_at IS NULL OR last_assigned_at < datetime('now', '-1 month')))
  );

-- name: UpdateTaskLastAssignedAt :exec
UPDATE tasks SET last_assigned_at = CURRENT_TIMESTAMP WHERE task_id = ?;

-- name: UpdateTaskRepeatConnection :exec
UPDATE tasks SET repeat_connection_id = ? WHERE task_id = ? AND manager_id = ?;

-- =====================
-- ASSIGNED TASK MANAGEMENT (manager)
-- =====================

-- name: DeleteAssignedTask :exec
DELETE FROM assigned_tasks WHERE assigned_task_id = ?;

-- name: UpdateAssignedTask :exec
UPDATE assigned_tasks SET
    points = ?, duration_minutes = ?, timer_days = ?, timer_hours = ?, timer_minutes = ?, due_time = ?,
    requires_image = ?, num_images_required = ?,
    requires_video = ?, num_videos_required = ?, requires_audio = ?, num_audio_required = ?, min_word_count = ?
WHERE assigned_task_id = ?;

-- name: UpdateAssignedTaskTemplate :exec
UPDATE tasks SET
    title = ?, description = ?
WHERE task_id = ?;
