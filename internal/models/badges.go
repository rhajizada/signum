package models

import "time"

// CreateBadgeRequest defines the payload for creating a badge.
type CreateBadgeRequest struct {
	Subject string `json:"subject"`
	Status  string `json:"status"`
	Color   string `json:"color"`
	Style   string `json:"style"`
} // @name CreateBadgeRequest

// PatchBadgeRequest defines the payload for patching a badge.
type PatchBadgeRequest struct {
	Subject *string `json:"subject"`
	Status  *string `json:"status"`
	Color   *string `json:"color"`
	Style   *string `json:"style"`
} // @name PatchBadgeRequest

// Badge defines the badge payload returned from the API.
type Badge struct {
	ID        string    `json:"id"`
	Subject   string    `json:"subject"`
	Status    string    `json:"status"`
	Color     string    `json:"color"`
	Style     string    `json:"style"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
} // @name Badge

// CreateBadgeResponse defines the response payload for badge creation.
type CreateBadgeResponse struct {
	Badge

	Token string `json:"token"`
} // @name CreateBadgeResponse
