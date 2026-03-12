# Task Timer Issue Documentation

## Problem Statement

~~When a task is assigned to a worker, it displays a countdown timer (days/hours/minutes). However, this timer **does not decrease over time** — it remains static until the page is manually refreshed or the user navigates away and back to the Buckets page.~~

~~The timer should be a live countdown that decreases in real-time as time passes toward the task's deadline.~~

## ✅ RESOLUTION (Completed)

**Fixed on**: 2026-03-12

The timer now updates in real-time via JavaScript countdown logic with visual urgency indicators and late submission tracking.

### Implementation Details

1. **Client-Side Countdown**: JavaScript function `updateCountdowns()` runs every second to calculate remaining time from `due_time`
2. **Data Attribute**: Timer elements include `data-due-time="{{ .DueTime }}"` to store the deadline timestamp
3. **Dynamic Formatting**: `formatTime()` converts milliseconds to readable format (e.g., "2d 5h 30m 15s")

### Visual Urgency System

**Timer Text Colors**:
- **Gray** (default): > 1 day remaining
- **Yellow + bold**: < 1 day remaining
- **Orange + bold**: < 1 hour remaining
- **Red + bold**: Expired

**Card Border Colors** (applies to To Do list items):
- **Emerald** (border-emerald-400): > 1 day remaining
- **Yellow** (border-yellow-400): < 1 day remaining
- **Orange** (border-orange-400): < 1 hour remaining
- **Red** (border-red-500): Expired

**Modal Border Colors** (applies when task detail is opened):
- Same color progression as cards
- Provides visual urgency when worker is viewing/submitting task

### Expired Task Handling

**In To Do section**:
- Expired tasks show red border and "Expired" timer
- When modal opened, displays warning banner:
  - Red background with alert icon
  - "Time Expired" heading
  - Message: "This task is overdue. You can still submit, but it will be marked as late."
- Worker can still submit expired tasks

**In Review section**:
- Late submissions display red clock icon next to points
- Modal shows "Submitted Late" warning banner
- Automatically detected by comparing current time to `due_time`

### Technical Details

**Data Attributes Used**:
- `data-due-time` — Timer text displays (countdown timers)
- `data-task-card` — To Do list item borders (urgency indication)
- `data-modal-task` — Modal borders and expired warnings (To Do section)
- `data-check-late` — Late submission indicators (Review section)

**JavaScript Functions**:
- `formatTime(ms)` — Converts milliseconds to human-readable time
- `updateCountdowns()` — Main update loop (runs every second)
  - Updates timer text
  - Updates card border colors
  - Updates modal border colors
  - Shows/hides expired/late warnings
  - Applies font weight changes (bold/semibold for urgency)

**Files Modified**:
- `internal/templates/fractions/buckets.html` — Added countdown JavaScript, urgency styling, warning banners, and data attributes

### Thresholds

- **1 day** (86400000ms): Yellow urgency starts
- **1 hour** (3600000ms): Orange urgency starts
- **0** (expired): Red urgency + warnings

---

## Current Implementation

### Database Schema

**`assigned_tasks` table** (database/schema.sql, lines 112-139):
- `assigned_at` TIMESTAMP — When the task was assigned (set to `CURRENT_TIMESTAMP`)
- `due_time` TIMESTAMP — Computed deadline (assigned_at + duration_minutes)
- `duration_minutes` INTEGER — Total time allowed for the task
- `timer_days`, `timer_hours`, `timer_minutes` — Nullable display fields (stored at assignment time)

**Key issue**: `timer_days`, `timer_hours`, `timer_minutes` are **static snapshots** stored in the database at assignment time. They never decrease.

### Server-Side (Go)

**File**: `internal/buckets/services.go`

**Function**: `calculateDurationMinutes()` (lines 16-32)
- Converts timer_days/hours/minutes into total duration_minutes
- Used during task creation/assignment to set the deadline
- Does NOT recalculate on retrieval

**Function**: `assignTaskToWorker()` (lines 220-260)
- Stores `duration_minutes` at assignment time
- Calculates `due_time = time.Now().Add(time.Duration(durationMinutes) * time.Minute)`
- Stores static `timer_days`, `timer_hours`, `timer_minutes` in `assigned_tasks`
- These values are **never updated** after assignment

**Query**: `GetAssignedTasksByConnectionAndStatus` (database/query.sql.go, line 624)
- Fetches task rows including the static `timer_days`, `timer_hours`, `timer_minutes`
- These are passed directly to the template for display

### Frontend (HTML/Templates)

**File**: `internal/templates/fractions/buckets.html`

**Lines 66-73**: Timer display on list items
```html
{{ if or .TimerDays.Valid .TimerHours.Valid .TimerMinutes.Valid }}
<div class="flex items-center gap-1 text-xs text-gray-400 mt-2">
    <svg><!-- clock icon --></svg>
    {{ if .TimerDays.Valid }}{{ .TimerDays.Int64 }}d {{ end }}
    {{ if .TimerHours.Valid }}{{ .TimerHours.Int64 }}h {{ end }}
    {{ if .TimerMinutes.Valid }}{{ .TimerMinutes.Int64 }}m{{ end }}
</div>
{{ end }}
```

**Lines 106-112**: Timer display in task modal (same pattern)
```html
{{ if or .TimerDays.Valid .TimerHours.Valid .TimerMinutes.Valid }}
<span class="flex items-center gap-1 text-gray-300">
    <svg><!-- clock icon --></svg>
    {{ if .TimerDays.Valid }}{{ .TimerDays.Int64 }}d {{ end }}
    {{ if .TimerHours.Valid }}{{ .TimerHours.Int64 }}h {{ end }}
    {{ if .TimerMinutes.Valid }}{{ .TimerMinutes.Int64 }}m{{ end }}
</span>
{{ end }}
```

**Issue**: These templates render **static values** from the database. There is **no JavaScript countdown logic** to decrement the timer in real-time.

---

## Root Causes

1. **Static Database Values**: `timer_days/hours/minutes` are snapshots captured at assignment time and never updated.

2. **No Client-Side Countdown**: The frontend renders these static values without any JavaScript to calculate remaining time based on `due_time` and current time.

3. **No Server-Side Updates**: There is no background job or scheduled task that recalculates and updates the timer fields after assignment.

4. **No Real-Time Updates**: No HTMX polling or WebSocket integration to refresh the timer display at intervals.

---

## What Needs to Change

### Option A: Client-Side Countdown (Recommended for simplicity)
- Pass `due_time` and `assigned_at` to the template instead of (or in addition to) the static timer fields
- Add JavaScript to calculate remaining time on page load: `remaining = due_time - currentTime`
- Use `setInterval()` to update the timer display every second/minute
- Handle expired tasks (when timer reaches 0)

### Option B: Server-Side Recalculation
- Modify `getTasksByStatus()` to recalculate timer values from `due_time` and current server time
- Update returned rows to have fresh `timer_days/hours/minutes` values
- Still requires client-side `setInterval()` for live countdown if live updates are desired

### Option C: Hybrid Approach
- Server provides `due_time` to frontend
- Frontend calculates and displays remaining time in JavaScript
- Optional: Add HTMX polling every minute to refresh from server (in case times drift)

---

## Affected Files

1. **database/schema.sql** (lines 112-139)
   - `assigned_tasks` table definition
   - Fields: `timer_days`, `timer_hours`, `timer_minutes`, `assigned_at`, `due_time`

2. **database/models.go**
   - SQLC-generated struct `GetAssignedTasksByConnectionAndStatusRow`
   - Contains `TimerDays`, `TimerHours`, `TimerMinutes` (nullable int64)

3. **database/query.sql.go** (auto-generated)
   - Query `GetAssignedTasksByConnectionAndStatus` (line 624)
   - Returns rows with static timer values

4. **internal/buckets/services.go**
   - `calculateDurationMinutes()` (lines 16-32)
   - `assignTaskToWorker()` (lines 220-260)
   - `createAndAssignTask()` (lines 262-310)
   - `getTasksByStatus()` (lines 196-207) — calls the query

5. **internal/templates/fractions/buckets.html**
   - Timer display on To Do list items (lines 66-73)
   - Timer display in task modal (lines 106-112)
   - Timer display in Saved Tasks form inputs (lines 467-476)

---

## Related Features

- **Task Expiration**: Query `GetExpiredTodoTasks` (database/query.sql.go, line 1140) checks `due_time <= CURRENT_TIMESTAMP` — this correctly uses the deadline, not the static timer fields.
- **Task Submission**: Submissions can be saved as drafts and submitted before due_time expires.
- **Auto-Submit on Timer Expiry**: Currently NOT implemented — no logic to auto-submit when timer hits 0.

---

## Testing Notes

- Create a task with a short duration (e.g., 5 minutes)
- Assign it to a worker
- Observe the timer on the Buckets page
- Wait 1 minute without refreshing the page
- Timer will NOT have decreased (BUG)
- Refresh the page — timer will show the updated value (if still within duration)

---

## Summary

The timer display is **purely static** — it's a snapshot of days/hours/minutes stored in the database when the task is assigned, rendered as-is in templates without any countdown logic. To fix this, the frontend needs JavaScript to calculate remaining time from the `due_time` field and update the display in real-time.
