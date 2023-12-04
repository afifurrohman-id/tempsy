package models

type ApiError struct {
	Type        string `json:"errorType"`
	Description string `json:"errorDescription"`
}
