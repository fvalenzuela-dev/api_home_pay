package models

import "time"

type User struct {
	AuthUserID string     `json:"auth_user_id"`
	FullName   string     `json:"full_name"`
	Email      string     `json:"email"`
	Timezone   string     `json:"timezone"`
	Currency   string     `json:"currency"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}
