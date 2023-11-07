-- Tasks

-- name: CreateTask :one
INSERT INTO "tasks" (
  name,
  description,
  task_type_id,
  case_id
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: ListTasks :many
SELECT * FROM "tasks";

-- name: GetTask :one
SELECT * FROM "tasks" WHERE id = $1;

-- name: CreateTaskType :one
INSERT INTO "task_types" (
  name
) VALUES (
  $1
) RETURNING *;

-- name: ListTaskTypes :many
SELECT * FROM "task_types";

-- name: GetTaskType :one
SELECT * FROM "task_types" WHERE id = $1;

-- name: CreateUserTask :one
INSERT INTO "user_tasks" (
  user_id,
  task_id,
  assigned_by,
  due_date,
  is_completed,
  reschedule_count
) VALUES (
  $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: ListUserTasks :many
SELECT * FROM "user_tasks";

-- name: GetUserTask :one
SELECT * FROM "user_tasks" WHERE id = $1;

-- name: GetUserTasksByUserId :many
SELECT * FROM "user_tasks" WHERE user_id = $1;

-- name: UpdateUserTask :one
UPDATE "user_tasks"
SET
  user_id = $2,
  task_id = $3,
  assigned_by = $4,
  due_date = $5,
  is_completed = $6,
  reschedule_count = $7
WHERE id = $1
RETURNING *;

-- name: DeleteUserTask :exec
DELETE FROM "user_tasks" WHERE id = $1;

-- name: UserTaskExists :one
SELECT EXISTS(SELECT 1 FROM "user_tasks" WHERE id = $1);

-- name: CreateTaskReschedule :one
INSERT INTO "task_reschedules" (
  user_task_id,
  new_due_date,
  reassigned_to,
  comment,
  rescheduled_by
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetTaskReschedule :one
SELECT * FROM "task_reschedules" WHERE id = $1;

-- name: ListTaskReschedules :many
SELECT * FROM "task_reschedules";

-- name: UpdateTaskReschedule :one
UPDATE "task_reschedules"
SET
  user_task_id = $2,
  new_due_date = $3,
  reassigned_to = $4,
  comment = $5,
  rescheduled_by = $6
WHERE id = $1
RETURNING *;

-- name: DeleteTaskReschedule :exec
DELETE FROM "task_reschedules" WHERE id = $1;

-- name: TaskRescheduleExists :one
SELECT EXISTS(SELECT 1 FROM "task_reschedules" WHERE id = $1);

-- name: CreateCalendarEvent :one
INSERT INTO "calendar_events" (
  id,
  user_id,
  case_id,
  event_date,
  notes,
  task_id
) VALUES (
  $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: ListCalendarEvents :many
SELECT * FROM "calendar_events";

-- name: GetCalendarEvent :one
SELECT * FROM "calendar_events" WHERE id = $1;

-- name: GetCalendarEventsByUserId :many
SELECT * FROM "calendar_events" WHERE user_id = $1;

