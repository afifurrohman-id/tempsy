package models

type DataFile struct {
	Name              string `json:"name"`
	AutoDeletedAt     int64  `json:"autoDeletedAt"`     // milliseconds
	PrivateUrlExpires int    `json:"privateUrlExpires"` // seconds
	IsPublic          bool   `json:"isPublic"`
	UploadedAt        int64  `json:"uploadedAt"` // milliseconds
	UpdatedAt         int64  `json:"updatedAt"`  // milliseconds
	Url               string `json:"url"`
	Size              int64  `json:"size"` // byte count
	ContentType       string `json:"type"`
}
