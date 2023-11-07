package service

import (
	"context"
	"fmt"
	"time"

	"github.com/miloszizic/der/db"

	"github.com/google/uuid"
)

type TaskType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type CreateTaskTypeParams struct {
	Name string `json:"name"`
}

// CreateTaskType creates a new task type in the database. It expects a name for the task type.
func (s *Stores) CreateTaskType(ctx context.Context, params CreateTaskTypeParams) (*TaskType, error) {
	taskType, err := s.DBStore.CreateTaskType(ctx, params.Name)
	if err != nil {
		return nil, fmt.Errorf("error creating task type: %w", err)
	}

	return &TaskType{
		ID:   taskType.ID.String(),
		Name: taskType.Name,
	}, nil
}

// GetTaskTypes returns a list of task types from the database.
func (s *Stores) GetTaskTypes(ctx context.Context) ([]*TaskType, error) {
	taskTypes, err := s.DBStore.ListTaskTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting task types: %w", err)
	}

	types := make([]*TaskType, 0, len(taskTypes))
	for _, t := range taskTypes {
		types = append(types, &TaskType{
			ID:   t.ID.String(),
			Name: t.Name,
		})
	}

	return types, nil
}

// GetTaskType returns a task type from the database. It expects the UUID of the task type.
func (s *Stores) GetTaskType(ctx context.Context, id uuid.UUID) (*TaskType, error) {
	taskType, err := s.DBStore.GetTaskType(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting task type: %w", err)
	}

	return &TaskType{
		ID:   taskType.ID.String(),
		Name: taskType.Name,
	}, nil
}

type Task struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	TaskTypeID  string        `json:"task_type_id"`
	CaseID      uuid.NullUUID `json:"case_id"`
}

type CreateTaskParams struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	TaskTypeID  uuid.UUID     `json:"task_type_id"`
	CaseID      uuid.NullUUID `json:"case_id"`
}

func convertDBTaskToServiceTask(t db.Task) *Task {
	var description string
	if t.Description.Valid {
		description = t.Description.String
	} else {
		description = ""
	}

	return &Task{
		ID:          t.ID.String(),
		Name:        t.Name,
		Description: description,
		TaskTypeID:  t.TaskTypeID.String(),
		CaseID:      t.CaseID,
	}
}

// CreateTask creates a new task in the database. It expects a name, description, task type ID, and case ID.
func (s *Stores) CreateTask(ctx context.Context, params CreateTaskParams) (*Task, error) {
	taskParams := db.CreateTaskParams{
		Name:        params.Name,
		Description: HandleNullableString(params.Description),
		TaskTypeID:  params.TaskTypeID,
		CaseID:      params.CaseID,
	}

	task, err := s.DBStore.CreateTask(ctx, taskParams)
	if err != nil {
		return nil, fmt.Errorf("error creating task: %w", err)
	}

	return convertDBTaskToServiceTask(task), nil
}

// GetTasks returns a list of tasks from the database.
func (s *Stores) GetTasks(ctx context.Context) ([]*Task, error) {
	tasks, err := s.DBStore.ListTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting tasks: %w", err)
	}

	ts := make([]*Task, 0, len(tasks))
	for _, t := range tasks {
		ts = append(ts, convertDBTaskToServiceTask(t))
	}

	return ts, nil
}

// GetTask returns a task from the database. It expects the UUID of the task.
func (s *Stores) GetTask(ctx context.Context, id uuid.UUID) (*Task, error) {
	task, err := s.DBStore.GetTask(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting task: %w", err)
	}

	return convertDBTaskToServiceTask(task), nil
}

type UserTask struct {
	ID              uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"user_id"`
	TaskID          uuid.UUID `json:"task_id"`
	AssignedBy      uuid.UUID `json:"assigned_by"`
	DueDate         time.Time `json:"due_date"`
	IsCompleted     bool      `json:"is_completed"`
	RescheduleCount int32     `json:"reschedule_count"`
}

type CreateUserTaskParams struct {
	UserID          uuid.UUID `json:"user_id"`
	TaskID          uuid.UUID `json:"task_id"`
	AssignedBy      uuid.UUID `json:"assigned_by"`
	DueDate         time.Time `json:"due_date"`
	IsCompleted     bool      `json:"is_completed"`
	RescheduleCount int32     `json:"reschedule_count"`
}

func convertDBUserTaskToServiceUserTask(ut db.UserTask) *UserTask {
	return &UserTask{
		ID:              ut.ID,
		UserID:          ut.UserID,
		TaskID:          ut.TaskID,
		AssignedBy:      ut.AssignedBy,
		DueDate:         ut.DueDate,
		IsCompleted:     ut.IsCompleted.Bool,
		RescheduleCount: ut.RescheduleCount.Int32,
	}
}

// CreateUserTask creates a new user task in the database. It expects a user ID and task ID.Optionally assigned by user ID, due date, is completed, and reschedule count.
func (s *Stores) CreateUserTask(ctx context.Context, params CreateUserTaskParams) (*UserTask, error) {
	userTaskParams := db.CreateUserTaskParams{
		UserID:          params.UserID,
		TaskID:          params.TaskID,
		AssignedBy:      params.AssignedBy,
		DueDate:         params.DueDate,
		IsCompleted:     HandleNullableBool(params.IsCompleted),
		RescheduleCount: HandleNullableInt32(params.RescheduleCount),
	}

	userTask, err := s.DBStore.CreateUserTask(ctx, userTaskParams)
	if err != nil {
		return nil, fmt.Errorf("error creating user task: %w", err)
	}

	return convertDBUserTaskToServiceUserTask(userTask), nil
}

// GetUserTasks returns a list of user tasks from the database.
func (s *Stores) GetUserTasks(ctx context.Context) ([]*UserTask, error) {
	userTasks, err := s.DBStore.ListUserTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting user tasks: %w", err)
	}

	uts := make([]*UserTask, 0, len(userTasks))
	for _, ut := range userTasks {
		uts = append(uts, convertDBUserTaskToServiceUserTask(ut))
	}

	return uts, nil
}

// GetUserTask returns a user task from the database. It expects the UUID of the user task.
func (s *Stores) GetUserTask(ctx context.Context, id uuid.UUID) (*UserTask, error) {
	userTask, err := s.DBStore.GetUserTask(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting user task: %w", err)
	}

	return convertDBUserTaskToServiceUserTask(userTask), nil
}

// GetTasksByUserID returns a list of user tasks from the database. It expects the UUID of the user.
func (s *Stores) GetTasksByUserID(ctx context.Context, userID uuid.UUID) ([]*UserTask, error) {
	userTasks, err := s.DBStore.GetUserTasksByUserId(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user tasks: %w", err)
	}

	uts := make([]*UserTask, 0, len(userTasks))
	for _, ut := range userTasks {
		uts = append(uts, convertDBUserTaskToServiceUserTask(ut))
	}

	return uts, nil
}

type CalendarEvent struct {
	ID        uuid.UUID     `json:"id"`
	UserID    uuid.UUID     `json:"user_id"`
	CaseID    uuid.UUID     `json:"case_id"`
	EventDate time.Time     `json:"event_date"`
	Notes     string        `json:"notes"`
	TaskID    uuid.NullUUID `json:"task_id"`
}

type CreateCalendarEventParams struct {
	UserID    uuid.UUID `json:"user_id"`
	CaseID    uuid.UUID `json:"case_id"`
	EventDate time.Time `json:"event_date"`
	Notes     string    `json:"notes"`
	TaskID    uuid.UUID `json:"task_id"`
}

func convertDBCalendarEventToServiceCalendarEvent(ce db.CalendarEvent) *CalendarEvent {
	return &CalendarEvent{
		ID:        ce.ID,
		UserID:    ce.UserID,
		CaseID:    ce.CaseID,
		EventDate: ce.EventDate,
		Notes:     ce.Notes.String,
		TaskID:    ce.TaskID,
	}
}

// CreateCalendarEvent creates a new calendar event in the database. It expects a user ID, case ID and event date.
func (s *Stores) CreateCalendarEvent(ctx context.Context, params CreateCalendarEventParams) (*CalendarEvent, error) {
	calendarEventParams := db.CreateCalendarEventParams{
		UserID:    params.UserID,
		CaseID:    params.CaseID,
		EventDate: params.EventDate,
		Notes:     HandleNullableString(params.Notes),
		TaskID:    HandleNullableUUID(params.TaskID),
	}

	calendarEvent, err := s.DBStore.CreateCalendarEvent(ctx, calendarEventParams)
	if err != nil {
		return nil, fmt.Errorf("error creating calendar event: %w", err)
	}

	return convertDBCalendarEventToServiceCalendarEvent(calendarEvent), nil
}

// GetCalendarEvents returns a list of calendar events from the database.
func (s *Stores) GetCalendarEvents(ctx context.Context) ([]*CalendarEvent, error) {
	calendarEvents, err := s.DBStore.ListCalendarEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting calendar events: %w", err)
	}

	ces := make([]*CalendarEvent, 0, len(calendarEvents))
	for _, ce := range calendarEvents {
		ces = append(ces, convertDBCalendarEventToServiceCalendarEvent(ce))
	}

	return ces, nil
}

// GetCalendarEvent returns a calendar event from the database. It expects the UUID of the calendar event.
func (s *Stores) GetCalendarEvent(ctx context.Context, id uuid.UUID) (*CalendarEvent, error) {
	calendarEvent, err := s.DBStore.GetCalendarEvent(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting calendar event: %w", err)
	}

	return convertDBCalendarEventToServiceCalendarEvent(calendarEvent), nil
}

// GetCalendarEventsByUserID returns a list of calendar events from the database. It expects the UUID of the user.
func (s *Stores) GetCalendarEventsByUserID(ctx context.Context, userID uuid.UUID) ([]*CalendarEvent, error) {
	calendarEvents, err := s.DBStore.GetCalendarEventsByUserId(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting calendar events: %w", err)
	}

	ces := make([]*CalendarEvent, 0, len(calendarEvents))
	for _, ce := range calendarEvents {
		ces = append(ces, convertDBCalendarEventToServiceCalendarEvent(ce))
	}

	return ces, nil
}
