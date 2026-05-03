package sessions

import "time"

type Session struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	UserAgent  string     `json:"user_agent,omitempty"`
	IPAddress  string     `json:"ip_address,omitempty"`
	ExpiresAt  time.Time  `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}
