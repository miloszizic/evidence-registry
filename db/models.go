// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1

package db

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type AppUser struct {
	ID        uuid.UUID      `json:"id"`
	Username  string         `json:"username"`
	Email     string         `json:"email"`
	Password  string         `json:"password"`
	RoleID    uuid.NullUUID  `json:"role_id"`
	FirstName sql.NullString `json:"first_name"`
	LastName  sql.NullString `json:"last_name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type AuditLog struct {
	ID        uuid.UUID      `json:"id"`
	Action    string         `json:"action"`
	TableName string         `json:"table_name"`
	RecordID  uuid.UUID      `json:"record_id"`
	OldData   sql.NullString `json:"old_data"`
	NewData   sql.NullString `json:"new_data"`
	ChangedAt time.Time      `json:"changed_at"`
	ChangedBy uuid.NullUUID  `json:"changed_by"`
}

type CalendarEvent struct {
	ID        uuid.UUID      `json:"id"`
	UserID    uuid.UUID      `json:"user_id"`
	CaseID    uuid.UUID      `json:"case_id"`
	EventDate time.Time      `json:"event_date"`
	Notes     sql.NullString `json:"notes"`
	TaskID    uuid.NullUUID  `json:"task_id"`
}

type Case struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `json:"name"`
	Tags        []string  `json:"tags"`
	CaseYear    int32     `json:"case_year"`
	CaseTypeID  uuid.UUID `json:"case_type_id"`
	CaseNumber  int32     `json:"case_number"`
	CaseCourtID uuid.UUID `json:"case_court_id"`
}

type CaseType struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type Court struct {
	ID        uuid.UUID `json:"id"`
	Code      int32     `json:"code"`
	Name      string    `json:"name"`
	ShortName string    `json:"short_name"`
}

type Evidence struct {
	ID             uuid.UUID      `json:"id"`
	CaseID         uuid.UUID      `json:"case_id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	AppUserID      uuid.UUID      `json:"app_user_id"`
	Name           string         `json:"name"`
	Description    sql.NullString `json:"description"`
	Hash           string         `json:"hash"`
	EvidenceTypeID uuid.UUID      `json:"evidence_type_id"`
}

type EvidenceType struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type Permission struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Code string    `json:"code"`
}

type Role struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Code string    `json:"code"`
}

type RolePermission struct {
	ID           uuid.UUID `json:"id"`
	RoleID       uuid.UUID `json:"role_id"`
	PermissionID uuid.UUID `json:"permission_id"`
}

type Session struct {
	ID               uuid.UUID `json:"id"`
	UserID           uuid.UUID `json:"user_id"`
	RefreshPayloadID uuid.UUID `json:"refresh_payload_id"`
	Username         string    `json:"username"`
	RefreshToken     string    `json:"refresh_token"`
	UserAgent        string    `json:"user_agent"`
	ClientIp         string    `json:"client_ip"`
	IsBlocked        bool      `json:"is_blocked"`
	ExpiresAt        time.Time `json:"expires_at"`
	CreatedAt        time.Time `json:"created_at"`
}

type SessionDatum struct {
	ID    uuid.UUID `json:"id"`
	Key   string    `json:"key"`
	Value uuid.UUID `json:"value"`
}

type Task struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	TaskTypeID  uuid.UUID      `json:"task_type_id"`
	CaseID      uuid.NullUUID  `json:"case_id"`
}

type TaskReschedule struct {
	ID            uuid.UUID      `json:"id"`
	UserTaskID    uuid.UUID      `json:"user_task_id"`
	NewDueDate    time.Time      `json:"new_due_date"`
	ReassignedTo  uuid.NullUUID  `json:"reassigned_to"`
	Comment       sql.NullString `json:"comment"`
	RescheduledBy uuid.UUID      `json:"rescheduled_by"`
}

type TaskType struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type UserCase struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	CaseID uuid.UUID `json:"case_id"`
}

type UserTask struct {
	ID              uuid.UUID     `json:"id"`
	UserID          uuid.UUID     `json:"user_id"`
	TaskID          uuid.UUID     `json:"task_id"`
	AssignedBy      uuid.UUID     `json:"assigned_by"`
	DueDate         time.Time     `json:"due_date"`
	IsCompleted     sql.NullBool  `json:"is_completed"`
	RescheduleCount sql.NullInt32 `json:"reschedule_count"`
}