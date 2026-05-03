package users

import "context"

type User struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

func CurrentUser(_ context.Context) User {
	return User{
		ID:          "00000000-0000-0000-0000-000000000000",
		Username:    "local-dev",
		Email:       "developer@clove.local",
		DisplayName: "Local Developer",
	}
}
