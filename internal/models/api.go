package models

type ApiError struct {
	Type        string `json:"error_type"`
	Description string `json:"error_description"`
}
