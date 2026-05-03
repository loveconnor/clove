package sshkeys

import "time"

type SSHKey struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Name        string     `json:"name"`
	PublicKey   string     `json:"public_key"`
	Fingerprint string     `json:"fingerprint"`
	CreatedAt   time.Time  `json:"created_at"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
}
