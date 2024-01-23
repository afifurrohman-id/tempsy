package models

type Error struct {
	Kind        string `json:"kind"`
	Description string `json:"description"`
}

type ApiError struct {
	*Error `json:"apiError"`
}
