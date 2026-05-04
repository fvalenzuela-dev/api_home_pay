package models

import "time"

type Category struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	AuthUserID string     `json:"auth_user_id"`
	IconWeb    *string    `json:"icon_web,omitempty"`
	IconApk    *string    `json:"icon_apk,omitempty"`
	ColorWeb   *string    `json:"color_web,omitempty"`
	ColorApk   *string    `json:"color_apk,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

type CreateCategoryRequest struct {
	Name      string `json:"name"`
	IconWeb   string `json:"icon_web"`
	IconApk   string `json:"icon_apk"`
	ColorWeb  string `json:"color_web"`
	ColorApk  string `json:"color_apk"`
}

type UpdateCategoryRequest struct {
	Name     *string `json:"name"`
	IconWeb  *string `json:"icon_web"`
	IconApk  *string `json:"icon_apk"`
	ColorWeb *string `json:"color_web"`
	ColorApk *string `json:"color_apk"`
}
