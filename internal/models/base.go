package models

import "time"

type Timestamp struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type UserContext struct {
	UserID string `json:"user_id"`
}
