package models

type DataFile struct {
	Name              string `json:"name"`
	AutoDeletedAt     int64  `json:"auto_deleted_at"`     // milliseconds
	PrivateUrlExpires int    `json:"private_url_expires"` // seconds
	IsPublic          bool   `json:"is_public"`
	UploadedAt        int64  `json:"uploaded_at"` // milliseconds
	UpdatedAt         int64  `json:"updated_at"`  // milliseconds
	Url               string `json:"url"`
	Size              int64  `json:"size"` // byte count
	ContentType       string `json:"type"`
}
