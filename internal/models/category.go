package models

import "time"

// Category represents a transaction category with platform-specific visual properties.
type Category struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	AuthUserID string     `json:"auth_user_id"`
	// IconWeb is the identifier or URL for the icon displayed on web platforms.
	IconWeb    *string    `json:"icon_web,omitempty"`
	// IconApk is the identifier or resource name for the icon displayed on mobile (APK) platforms.
	IconApk    *string    `json:"icon_apk,omitempty"`
	// ColorWeb is the hex color code or identifier used for the web platform.
	ColorWeb   *string    `json:"color_web,omitempty"`
	// ColorApk is the hex color code or identifier used for the mobile (APK) platform.
	ColorApk   *string    `json:"color_apk,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

// CreateCategoryRequest defines the payload for creating a new category.
// The Icon and Color fields are optional; they will be stored as NULL in the database if empty strings are provided.
type CreateCategoryRequest struct {
	Name     string `json:"name"`
	// IconWeb is the identifier or URL for the icon displayed on web platforms.
	IconWeb  string `json:"icon_web"`
	// IconApk is the identifier or resource name for the icon displayed on mobile (APK) platforms.
	IconApk  string `json:"icon_apk"`
	// ColorWeb is the hex color code or identifier used for the web platform.
	ColorWeb string `json:"color_web"`
	// ColorApk is the hex color code or identifier used for the mobile (APK) platform.
	ColorApk string `json:"color_apk"`
}

// UpdateCategoryRequest defines the payload for updating an existing category's fields.
// Pointers are used to allow partial updates; fields that are nil in the request will remain unchanged.
type UpdateCategoryRequest struct {
	Name     *string `json:"name"`
	// IconWeb is the identifier or URL for the icon displayed on web platforms.
	IconWeb  *string `json:"icon_web"`
	// IconApk is the identifier or resource name for the icon displayed on mobile (APK) platforms.
	IconApk  *string `json:"icon_apk"`
	// ColorWeb is the hex color code or identifier used for the web platform.
	ColorWeb *string `json:"color_web"`
	// ColorApk is the hex color code or identifier used for the mobile (APK) platform.
	ColorApk *string `json:"color_apk"`
}