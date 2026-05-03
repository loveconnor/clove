package auditlogs

import "time"

type AuditLog struct {
	ID             string         `json:"id"`
	ActorUserID    *string        `json:"actor_user_id,omitempty"`
	OrganizationID *string        `json:"organization_id,omitempty"`
	RepositoryID   *string        `json:"repository_id,omitempty"`
	Action         string         `json:"action"`
	TargetType     string         `json:"target_type,omitempty"`
	TargetID       *string        `json:"target_id,omitempty"`
	Metadata       map[string]any `json:"metadata"`
	CreatedAt      time.Time      `json:"created_at"`
}
