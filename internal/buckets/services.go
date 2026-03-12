package buckets

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pact/database"
	"strings"
	"time"
)

// Helper function to calculate duration in minutes from timer fields
func calculateDurationMinutes(timerDays, timerHours, timerMinutes sql.NullInt64) int64 {
	var total int64 = 0
	if timerDays.Valid {
		total += timerDays.Int64 * 1440 // days to minutes
	}
	if timerHours.Valid {
		total += timerHours.Int64 * 60 // hours to minutes
	}
	if timerMinutes.Valid {
		total += timerMinutes.Int64
	}
	// Default to 1440 minutes (1 day) if all fields are null/zero
	if total == 0 {
		total = 1440
	}
	return total
}

type RenderFunc func(http.ResponseWriter, string, interface{})

var renderFraction RenderFunc

func SetRenderFunc(fn RenderFunc) {
	renderFraction = fn
}

func renderBuckets(w http.ResponseWriter, r *http.Request) {
	data := BuildBucketsData(r)
	renderFraction(w, "buckets", data)
}

type BucketsData struct {
	Data map[string]any
}

func (b BucketsData) GetData() map[string]any {
	return b.Data
}

func BuildBucketsData(r *http.Request) BucketsData {
	userId := r.Context().Value("userID").(int)
	if userId < 1 {
		return BucketsData{
			Data: map[string]any{
				"Title":        "Buckets",
				"NoConnection": true,
			},
		}
	}

	connectionId, err := getActiveConnectionId(r.Context(), userId)
	if err != nil {
		return BucketsData{
			Data: map[string]any{
				"Title":        "Buckets",
				"NoConnection": true,
			},
		}
	}

	conn, err := getConnectionForBuckets(r.Context(), connectionId)
	if err != nil {
		return BucketsData{
			Data: map[string]any{
				"Title":        "Buckets",
				"NoConnection": true,
			},
		}
	}

	role := getUserRole(userId, conn)

	todoTasks, _ := getTasksByStatus(r.Context(), connectionId, "todo")
	reviewTasks, _ := getTasksByStatus(r.Context(), connectionId, "in_review")
	completedTasks, _ := getTasksByStatus(r.Context(), connectionId, "completed")
	failedTasks, _ := getTasksByStatus(r.Context(), connectionId, "failed")

	// Calculate overflow counts
	todoOverflow := 0
	if len(todoTasks) > 5 {
		todoOverflow = len(todoTasks) - 5
	}
	reviewOverflow := 0
	if len(reviewTasks) > 5 {
		reviewOverflow = len(reviewTasks) - 5
	}
	completedOverflow := 0
	if len(completedTasks) > 5 {
		completedOverflow = len(completedTasks) - 5
	}
	failedOverflow := 0
	if len(failedTasks) > 5 {
		failedOverflow = len(failedTasks) - 5
	}

	data := BucketsData{
		Data: map[string]any{
			"Title":             "Buckets",
			"Role":              role,
			"ConnectionId":      conn.ConnectionID,
			"WorkerPoints":      conn.WorkerPoints,
			"ManagerUsername":   conn.ManagerUsername,
			"WorkerUsername":    conn.WorkerUsername,
			"TodoTasks":         todoTasks,
			"ReviewTasks":       reviewTasks,
			"CompletedTasks":    completedTasks,
			"FailedTasks":       failedTasks,
			"TodoOverflow":      todoOverflow,
			"ReviewOverflow":    reviewOverflow,
			"CompletedOverflow": completedOverflow,
			"FailedOverflow":    failedOverflow,
		},
	}

	if role == "worker" {
		rewards, err := getRewardsForConnection(r.Context(), connectionId)
		if err == nil {
			data.Data["Rewards"] = rewards
			rewardsOverflow := 0
			if len(rewards) > 5 {
				rewardsOverflow = len(rewards) - 5
			}
			data.Data["RewardsOverflow"] = rewardsOverflow
		}
	}

	if role == "manager" {
		rewards, err := getRewardsForConnection(r.Context(), connectionId)
		if err == nil {
			data.Data["Rewards"] = rewards
			rewardsOverflow := 0
			if len(rewards) > 5 {
				rewardsOverflow = len(rewards) - 5
			}
			data.Data["RewardsOverflow"] = rewardsOverflow
		}
		saved, err := getBookmarkedTasks(r.Context(), int64(userId))
		if err == nil {
			data.Data["SavedTasks"] = saved
			savedOverflow := 0
			if len(saved) > 5 {
				savedOverflow = len(saved) - 5
			}
			data.Data["SavedOverflow"] = savedOverflow
		}
	}

	return data
}

func getConnectionForBuckets(ctx context.Context, connectionId int64) (database.GetConnectionForBucketsRow, error) {
	queries := database.GetQueries()

	conn, err := queries.GetConnectionForBuckets(ctx, connectionId)
	if err != nil {
		return database.GetConnectionForBucketsRow{}, fmt.Errorf("could not get connection %d: %w", connectionId, err)
	}
	return conn, nil
}

func getActiveConnectionId(ctx context.Context, userId int) (int64, error) {
	queries := database.GetQueries()

	connId, err := queries.GetActiveConnectionId(ctx, int64(userId))
	if err != nil {
		return 0, fmt.Errorf("could not get active connection for user %d: %w", userId, err)
	}
	if !connId.Valid {
		return 0, fmt.Errorf("no active connection set")
	}
	return connId.Int64, nil
}

func getUserRole(userId int, conn database.GetConnectionForBucketsRow) string {
	if int64(userId) == conn.ManagerID {
		return "manager"
	}
	return "worker"
}

func getTasksByStatus(ctx context.Context, connectionId int64, status string) ([]database.GetAssignedTasksByConnectionAndStatusRow, error) {
	queries := database.GetQueries()

	rows, err := queries.GetAssignedTasksByConnectionAndStatus(ctx, database.GetAssignedTasksByConnectionAndStatusParams{
		ConnectionID: connectionId,
		Status:       status,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get %s tasks for connection %d: %w", status, connectionId, err)
	}
	return rows, nil
}

func createTaskTemplate(ctx context.Context, managerId int64, params database.CreateTaskParams) (int64, error) {
	queries := database.GetQueries()

	params.ManagerID = managerId
	taskId, err := queries.CreateTask(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("could not create task template: %w", err)
	}
	return taskId, nil
}

func assignTaskToWorker(ctx context.Context, managerId int64, taskId int64, connectionId int64, durationMinutes int64, points int64, dueTime time.Time) (int64, error) {
	queries := database.GetQueries()

	task, err := queries.GetTaskById(ctx, taskId)
	if err != nil {
		return 0, fmt.Errorf("could not find task template %d: %w", taskId, err)
	}
	if task.ManagerID != managerId {
		return 0, fmt.Errorf("task %d does not belong to manager %d", taskId, managerId)
	}

	conn, err := queries.GetConnectionForBuckets(ctx, connectionId)
	if err != nil {
		return 0, fmt.Errorf("could not find connection %d: %w", connectionId, err)
	}
	if conn.ManagerID != managerId {
		return 0, fmt.Errorf("user %d is not the manager of connection %d", managerId, connectionId)
	}

	assignedId, err := queries.AssignTask(ctx, database.AssignTaskParams{
		TaskID:            taskId,
		ConnectionID:      connectionId,
		WorkerID:          conn.WorkerID,
		Type:              task.Type,
		Points:            points,
		DurationMinutes:   durationMinutes,
		DueTime:           dueTime,
		RequiresImage:     task.RequiresImage,
		NumImagesRequired: task.NumImagesRequired,
		RequiresVideo:     task.RequiresVideo,
		NumVideosRequired: task.NumVideosRequired,
		RequiresAudio:     task.RequiresAudio,
		NumAudioRequired:  task.NumAudioRequired,
		MinWordCount:      task.MinWordCount,
		PunishmentTaskID:  task.PunishmentTaskID,
	})
	if err != nil {
		return 0, fmt.Errorf("could not assign task: %w", err)
	}
	return assignedId, nil
}

func createAndAssignTask(ctx context.Context, managerId int64, connectionId int64, taskParams database.CreateTaskParams, durationMinutes int64, points int64, timerDays, timerHours, timerMinutes sql.NullInt64) (int64, error) {
	db := database.GetDB()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := database.New(tx)

	conn, err := qtx.GetConnectionForBuckets(ctx, connectionId)
	if err != nil {
		return 0, fmt.Errorf("could not find connection %d: %w", connectionId, err)
	}
	if conn.ManagerID != managerId {
		return 0, fmt.Errorf("user %d is not the manager of connection %d", managerId, connectionId)
	}

	taskParams.ManagerID = managerId
	taskId, err := qtx.CreateTask(ctx, taskParams)
	if err != nil {
		return 0, fmt.Errorf("could not create task: %w", err)
	}

	dueTime := time.Now().Add(time.Duration(durationMinutes) * time.Minute)

	assignedId, err := qtx.AssignTask(ctx, database.AssignTaskParams{
		TaskID:            taskId,
		ConnectionID:      connectionId,
		WorkerID:          conn.WorkerID,
		Type:              taskParams.Type,
		Points:            points,
		DurationMinutes:   durationMinutes,
		TimerDays:         timerDays,
		TimerHours:        timerHours,
		TimerMinutes:      timerMinutes,
		DueTime:           dueTime,
		RequiresImage:     taskParams.RequiresImage,
		NumImagesRequired: taskParams.NumImagesRequired,
		RequiresVideo:     taskParams.RequiresVideo,
		NumVideosRequired: taskParams.NumVideosRequired,
		RequiresAudio:     taskParams.RequiresAudio,
		NumAudioRequired:  taskParams.NumAudioRequired,
		MinWordCount:      taskParams.MinWordCount,
		PunishmentTaskID:  taskParams.PunishmentTaskID,
	})
	if err != nil {
		return 0, fmt.Errorf("could not assign task: %w", err)
	}

	if taskParams.RepeatFrequency.Valid && taskParams.RepeatFrequency.String != "" {
		err = qtx.UpdateTaskLastAssignedAt(ctx, taskId)
		if err != nil {
			return 0, fmt.Errorf("could not update last_assigned_at: %w", err)
		}
		err = qtx.UpdateTaskRepeatConnection(ctx, database.UpdateTaskRepeatConnectionParams{
			RepeatConnectionID: sql.NullInt64{Int64: connectionId, Valid: true},
			TaskID:             taskId,
			ManagerID:          managerId,
		})
		if err != nil {
			return 0, fmt.Errorf("could not set repeat connection: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("could not commit transaction: %w", err)
	}

	return assignedId, nil
}

func saveSubmission(ctx context.Context, workerId int64, assignedTaskId int64, submissionText string) error {
	queries := database.GetQueries()

	task, err := queries.GetAssignedTaskById(ctx, assignedTaskId)
	if err != nil {
		return fmt.Errorf("could not find assigned task %d: %w", assignedTaskId, err)
	}
	if task.WorkerID != workerId {
		return fmt.Errorf("task %d does not belong to worker %d", assignedTaskId, workerId)
	}
	if task.Status != "todo" {
		return fmt.Errorf("task %d is not in todo status", assignedTaskId)
	}

	_, err = queries.UpsertSubmission(ctx, database.UpsertSubmissionParams{
		AssignedTaskID: assignedTaskId,
		SubmissionText: sql.NullString{String: submissionText, Valid: submissionText != ""},
	})
	if err != nil {
		return fmt.Errorf("could not save submission: %w", err)
	}
	return nil
}

func submitTask(ctx context.Context, workerId int64, assignedTaskId int64, submissionText string, imagePathsJSON string, videoPathsJSON string, audioPathsJSON string) error {
	db := database.GetDB()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := database.New(tx)

	task, err := qtx.GetAssignedTaskById(ctx, assignedTaskId)
	if err != nil {
		return fmt.Errorf("could not find assigned task %d: %w", assignedTaskId, err)
	}
	if task.WorkerID != workerId {
		return fmt.Errorf("task %d does not belong to worker %d", assignedTaskId, workerId)
	}
	if task.Status != "todo" {
		return fmt.Errorf("task %d is not in todo status", assignedTaskId)
	}

	if task.RequiresImage == 1 {
		var paths []string
		if imagePathsJSON != "" && imagePathsJSON != "null" {
			json.Unmarshal([]byte(imagePathsJSON), &paths)
		}
		if int64(len(paths)) < task.NumImagesRequired {
			return fmt.Errorf("need %d image(s), have %d", task.NumImagesRequired, len(paths))
		}
	}
	if task.RequiresVideo == 1 {
		var paths []string
		if videoPathsJSON != "" && videoPathsJSON != "null" {
			json.Unmarshal([]byte(videoPathsJSON), &paths)
		}
		if int64(len(paths)) < task.NumVideosRequired {
			return fmt.Errorf("need %d video(s), have %d", task.NumVideosRequired, len(paths))
		}
	}
	if task.RequiresAudio == 1 {
		var paths []string
		if audioPathsJSON != "" && audioPathsJSON != "null" {
			json.Unmarshal([]byte(audioPathsJSON), &paths)
		}
		if int64(len(paths)) < task.NumAudioRequired {
			return fmt.Errorf("need %d audio file(s), have %d", task.NumAudioRequired, len(paths))
		}
	}
	if task.MinWordCount.Valid && task.MinWordCount.Int64 > 0 {
		wordCount := len(strings.Fields(submissionText))
		if int64(wordCount) < task.MinWordCount.Int64 {
			return fmt.Errorf("submission must have at least %d words (currently %d)", task.MinWordCount.Int64, wordCount)
		}
	}

	_, err = qtx.UpsertSubmission(ctx, database.UpsertSubmissionParams{
		AssignedTaskID: assignedTaskId,
		SubmissionText: sql.NullString{String: submissionText, Valid: submissionText != ""},
		ImagePaths:     sql.NullString{String: imagePathsJSON, Valid: imagePathsJSON != "" && imagePathsJSON != "null"},
		VideoPaths:     sql.NullString{String: videoPathsJSON, Valid: videoPathsJSON != "" && videoPathsJSON != "null"},
		AudioPaths:     sql.NullString{String: audioPathsJSON, Valid: audioPathsJSON != "" && audioPathsJSON != "null"},
	})
	if err != nil {
		return fmt.Errorf("could not save submission: %w", err)
	}

	err = qtx.SubmitAssignedTask(ctx, assignedTaskId)
	if err != nil {
		return fmt.Errorf("could not submit task: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}
	return nil
}

func approveTask(ctx context.Context, managerId int64, assignedTaskId int64) error {
	db := database.GetDB()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := database.New(tx)

	task, err := qtx.GetAssignedTaskById(ctx, assignedTaskId)
	if err != nil {
		return fmt.Errorf("could not find assigned task %d: %w", assignedTaskId, err)
	}
	if task.Status != "in_review" {
		return fmt.Errorf("task %d is not in review", assignedTaskId)
	}

	conn, err := qtx.GetConnectionForBuckets(ctx, task.ConnectionID)
	if err != nil {
		return fmt.Errorf("could not find connection: %w", err)
	}
	if conn.ManagerID != managerId {
		return fmt.Errorf("user %d is not the manager", managerId)
	}

	err = qtx.CompleteAssignedTask(ctx, assignedTaskId)
	if err != nil {
		return fmt.Errorf("could not complete task: %w", err)
	}

	err = qtx.AddWorkerPoints(ctx, database.AddWorkerPointsParams{
		WorkerPoints: task.Points,
		ConnectionID: task.ConnectionID,
	})
	if err != nil {
		return fmt.Errorf("could not add points: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}
	return nil
}

func disapproveTask(ctx context.Context, managerId int64, assignedTaskId int64) error {
	db := database.GetDB()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := database.New(tx)

	task, err := qtx.GetAssignedTaskById(ctx, assignedTaskId)
	if err != nil {
		return fmt.Errorf("could not find assigned task %d: %w", assignedTaskId, err)
	}
	if task.Status != "in_review" {
		return fmt.Errorf("task %d is not in review", assignedTaskId)
	}

	conn, err := qtx.GetConnectionForBuckets(ctx, task.ConnectionID)
	if err != nil {
		return fmt.Errorf("could not find connection: %w", err)
	}
	if conn.ManagerID != managerId {
		return fmt.Errorf("user %d is not the manager", managerId)
	}

	err = qtx.FailAssignedTask(ctx, assignedTaskId)
	if err != nil {
		return fmt.Errorf("could not fail task: %w", err)
	}

	if task.PunishmentTaskID.Valid {
		punishmentTask, err := qtx.GetTaskById(ctx, task.PunishmentTaskID.Int64)
		if err == nil {
			dueTime := time.Now().Add(time.Duration(punishmentTask.DefaultDurationMinutes) * time.Minute)
			_, err = qtx.AssignTask(ctx, database.AssignTaskParams{
				TaskID:            punishmentTask.TaskID,
				ConnectionID:      task.ConnectionID,
				WorkerID:          task.WorkerID,
				Type:              "punishment",
				Points:            punishmentTask.DefaultPoints,
				DurationMinutes:   punishmentTask.DefaultDurationMinutes,
				DueTime:           dueTime,
				RequiresImage:     punishmentTask.RequiresImage,
				NumImagesRequired: punishmentTask.NumImagesRequired,
				RequiresVideo:     punishmentTask.RequiresVideo,
				NumVideosRequired: punishmentTask.NumVideosRequired,
				RequiresAudio:     punishmentTask.RequiresAudio,
				NumAudioRequired:  punishmentTask.NumAudioRequired,
				MinWordCount:      punishmentTask.MinWordCount,
				PunishmentTaskID:  punishmentTask.PunishmentTaskID,
			})
			if err != nil {
				return fmt.Errorf("could not assign punishment task: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}
	return nil
}

func purchaseReward(ctx context.Context, workerId int64, connectionId int64, taskId int64) (int64, error) {
	db := database.GetDB()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := database.New(tx)

	conn, err := qtx.GetConnectionForBuckets(ctx, connectionId)
	if err != nil {
		return 0, fmt.Errorf("could not find connection %d: %w", connectionId, err)
	}
	if conn.WorkerID != workerId {
		return 0, fmt.Errorf("user %d is not the worker of connection %d", workerId, connectionId)
	}

	task, err := qtx.GetTaskById(ctx, taskId)
	if err != nil {
		return 0, fmt.Errorf("could not find reward task %d: %w", taskId, err)
	}
	if task.Type != "reward" {
		return 0, fmt.Errorf("task %d is not a reward", taskId)
	}
	if task.ManagerID != conn.ManagerID {
		return 0, fmt.Errorf("reward does not belong to this connection's manager")
	}

	cost := task.PointCost.Int64
	if !task.PointCost.Valid || cost <= 0 {
		return 0, fmt.Errorf("reward has no valid point cost")
	}
	if conn.WorkerPoints < cost {
		return 0, fmt.Errorf("not enough points: have %d, need %d", conn.WorkerPoints, cost)
	}

	err = qtx.DeductWorkerPoints(ctx, database.DeductWorkerPointsParams{
		WorkerPoints: cost,
		ConnectionID: connectionId,
	})
	if err != nil {
		return 0, fmt.Errorf("could not deduct points: %w", err)
	}

	dueTime := time.Now().Add(time.Duration(task.DefaultDurationMinutes) * time.Minute)
	assignedId, err := qtx.AssignTask(ctx, database.AssignTaskParams{
		TaskID:          taskId,
		ConnectionID:    connectionId,
		WorkerID:        workerId,
		Type:            "reward",
		Points:          task.DefaultPoints,
		DurationMinutes: task.DefaultDurationMinutes,
		DueTime:         dueTime,
		RequiresImage:   task.RequiresImage,
		RequiresVideo:   task.RequiresVideo,
		RequiresAudio:   task.RequiresAudio,
		MinWordCount:    task.MinWordCount,
	})
	if err != nil {
		return 0, fmt.Errorf("could not assign reward task: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("could not commit transaction: %w", err)
	}
	return assignedId, nil
}

func deleteTaskTemplate(ctx context.Context, managerId int64, taskId int64) error {
	queries := database.GetQueries()

	err := queries.DeleteTask(ctx, database.DeleteTaskParams{
		TaskID:    taskId,
		ManagerID: managerId,
	})
	if err != nil {
		return fmt.Errorf("could not delete task %d: %w", taskId, err)
	}
	return nil
}

func updateTaskTemplate(ctx context.Context, managerId int64, taskId int64, params database.UpdateTaskParams) error {
	queries := database.GetQueries()

	params.TaskID = taskId
	params.ManagerID = managerId

	err := queries.UpdateTask(ctx, params)
	if err != nil {
		return fmt.Errorf("could not update task %d: %w", taskId, err)
	}
	return nil
}

func getRewardsForConnection(ctx context.Context, connectionId int64) ([]database.GetAvailableRewardsRow, error) {
	queries := database.GetQueries()

	rows, err := queries.GetAvailableRewards(ctx, connectionId)
	if err != nil {
		return nil, fmt.Errorf("could not get rewards for connection %d: %w", connectionId, err)
	}
	return rows, nil
}

func getWorkerPoints(ctx context.Context, connectionId int64) (int64, error) {
	queries := database.GetQueries()

	points, err := queries.GetWorkerPoints(ctx, connectionId)
	if err != nil {
		return 0, fmt.Errorf("could not get worker points: %w", err)
	}
	return points, nil
}

func getBookmarkedTasks(ctx context.Context, managerId int64) ([]database.Task, error) {
	queries := database.GetQueries()

	tasks, err := queries.GetBookmarkedTasks(ctx, managerId)
	if err != nil {
		return nil, fmt.Errorf("could not get bookmarked tasks for manager %d: %w", managerId, err)
	}
	return tasks, nil
}

func assignFromTemplate(ctx context.Context, managerId int64, taskId int64, connectionId int64, durationMinutes int64, points int64, timerDays, timerHours, timerMinutes sql.NullInt64) (int64, error) {
	db := database.GetDB()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := database.New(tx)

	task, err := qtx.GetTaskById(ctx, taskId)
	if err != nil {
		return 0, fmt.Errorf("could not find task template %d: %w", taskId, err)
	}
	if task.ManagerID != managerId {
		return 0, fmt.Errorf("task %d does not belong to manager %d", taskId, managerId)
	}

	conn, err := qtx.GetConnectionForBuckets(ctx, connectionId)
	if err != nil {
		return 0, fmt.Errorf("could not find connection %d: %w", connectionId, err)
	}
	if conn.ManagerID != managerId {
		return 0, fmt.Errorf("user %d is not the manager of connection %d", managerId, connectionId)
	}

	dueTime := time.Now().Add(time.Duration(durationMinutes) * time.Minute)

	assignedId, err := qtx.AssignTask(ctx, database.AssignTaskParams{
		TaskID:            taskId,
		ConnectionID:      connectionId,
		WorkerID:          conn.WorkerID,
		Type:              task.Type,
		Points:            points,
		DurationMinutes:   durationMinutes,
		TimerDays:         timerDays,
		TimerHours:        timerHours,
		TimerMinutes:      timerMinutes,
		DueTime:           dueTime,
		RequiresImage:     task.RequiresImage,
		NumImagesRequired: task.NumImagesRequired,
		RequiresVideo:     task.RequiresVideo,
		NumVideosRequired: task.NumVideosRequired,
		RequiresAudio:     task.RequiresAudio,
		NumAudioRequired:  task.NumAudioRequired,
		MinWordCount:      task.MinWordCount,
		PunishmentTaskID:  task.PunishmentTaskID,
	})
	if err != nil {
		return 0, fmt.Errorf("could not assign task: %w", err)
	}

	if task.RepeatFrequency.Valid && task.RepeatFrequency.String != "" {
		err = qtx.UpdateTaskLastAssignedAt(ctx, taskId)
		if err != nil {
			return 0, fmt.Errorf("could not update last_assigned_at: %w", err)
		}
		if !task.RepeatConnectionID.Valid {
			err = qtx.UpdateTaskRepeatConnection(ctx, database.UpdateTaskRepeatConnectionParams{
				RepeatConnectionID: sql.NullInt64{Int64: connectionId, Valid: true},
				TaskID:             taskId,
				ManagerID:          managerId,
			})
			if err != nil {
				return 0, fmt.Errorf("could not set repeat connection: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("could not commit transaction: %w", err)
	}
	return assignedId, nil
}

func deleteAssignedTask(ctx context.Context, managerId int64, assignedTaskId int64) error {
	queries := database.GetQueries()

	task, err := queries.GetAssignedTaskById(ctx, assignedTaskId)
	if err != nil {
		return fmt.Errorf("could not find assigned task %d: %w", assignedTaskId, err)
	}

	conn, err := queries.GetConnectionForBuckets(ctx, task.ConnectionID)
	if err != nil {
		return fmt.Errorf("could not find connection: %w", err)
	}
	if conn.ManagerID != managerId {
		return fmt.Errorf("user %d is not the manager", managerId)
	}

	err = queries.DeleteAssignedTask(ctx, assignedTaskId)
	if err != nil {
		return fmt.Errorf("could not delete assigned task %d: %w", assignedTaskId, err)
	}
	return nil
}

func updateAssignedTask(ctx context.Context, managerId int64, assignedTaskId int64, title string, description sql.NullString, points int64, durationMinutes int64, timerDays, timerHours, timerMinutes sql.NullInt64, requiresImage int64, numImagesRequired int64, requiresVideo int64, numVideosRequired int64, requiresAudio int64, numAudioRequired int64, minWordCount sql.NullInt64) error {
	db := database.GetDB()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := database.New(tx)

	task, err := qtx.GetAssignedTaskById(ctx, assignedTaskId)
	if err != nil {
		return fmt.Errorf("could not find assigned task %d: %w", assignedTaskId, err)
	}

	conn, err := qtx.GetConnectionForBuckets(ctx, task.ConnectionID)
	if err != nil {
		return fmt.Errorf("could not find connection: %w", err)
	}
	if conn.ManagerID != managerId {
		return fmt.Errorf("user %d is not the manager", managerId)
	}

	if task.Status != "todo" {
		return fmt.Errorf("can only edit tasks in todo status")
	}

	// Calculate new due time from assigned_at if available
	var dueTime time.Time
	if task.AssignedAt.Valid {
		dueTime = task.AssignedAt.Time.Add(time.Duration(durationMinutes) * time.Minute)
	} else {
		dueTime = time.Now().Add(time.Duration(durationMinutes) * time.Minute)
	}

	// Update the task template (title, description)
	err = qtx.UpdateAssignedTaskTemplate(ctx, database.UpdateAssignedTaskTemplateParams{
		Title:       title,
		Description: description,
		TaskID:      task.TaskID,
	})
	if err != nil {
		return fmt.Errorf("could not update task template: %w", err)
	}

	// Update the assigned task instance
	err = qtx.UpdateAssignedTask(ctx, database.UpdateAssignedTaskParams{
		Points:            points,
		DurationMinutes:   durationMinutes,
		TimerDays:         timerDays,
		TimerHours:        timerHours,
		TimerMinutes:      timerMinutes,
		DueTime:           dueTime,
		RequiresImage:     requiresImage,
		NumImagesRequired: numImagesRequired,
		RequiresVideo:     requiresVideo,
		NumVideosRequired: numVideosRequired,
		RequiresAudio:     requiresAudio,
		NumAudioRequired:  numAudioRequired,
		MinWordCount:      minWordCount,
		AssignedTaskID:    assignedTaskId,
	})
	if err != nil {
		return fmt.Errorf("could not update assigned task: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}
	return nil
}

func ProcessDueRepeatingTasks() {
	db := database.GetDB()
	ctx := context.Background()
	queries := database.GetQueries()

	dueTasks, err := queries.GetAllDueRepeatingTasks(ctx)
	if err != nil {
		log.Printf("error getting due repeating tasks: %v", err)
		return
	}

	for _, task := range dueTasks {
		if !task.RepeatConnectionID.Valid {
			continue
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			log.Printf("error beginning tx for repeat task %d: %v", task.TaskID, err)
			continue
		}

		qtx := database.New(tx)

		conn, err := qtx.GetConnectionForBuckets(ctx, task.RepeatConnectionID.Int64)
		if err != nil {
			tx.Rollback()
			log.Printf("error getting connection for repeat task %d: %v", task.TaskID, err)
			continue
		}

		dueTime := time.Now().Add(time.Duration(task.DefaultDurationMinutes) * time.Minute)

		_, err = qtx.AssignTask(ctx, database.AssignTaskParams{
			TaskID:           task.TaskID,
			ConnectionID:     task.RepeatConnectionID.Int64,
			WorkerID:         conn.WorkerID,
			Type:             task.Type,
			Points:           task.DefaultPoints,
			DurationMinutes:  task.DefaultDurationMinutes,
			DueTime:          dueTime,
			RequiresImage:    task.RequiresImage,
			RequiresVideo:    task.RequiresVideo,
			RequiresAudio:    task.RequiresAudio,
			MinWordCount:     task.MinWordCount,
			PunishmentTaskID: task.PunishmentTaskID,
		})
		if err != nil {
			tx.Rollback()
			log.Printf("error assigning repeat task %d: %v", task.TaskID, err)
			continue
		}

		err = qtx.UpdateTaskLastAssignedAt(ctx, task.TaskID)
		if err != nil {
			tx.Rollback()
			log.Printf("error updating last_assigned_at for task %d: %v", task.TaskID, err)
			continue
		}

		if err = tx.Commit(); err != nil {
			log.Printf("error committing repeat task %d: %v", task.TaskID, err)
		}
	}
}
