package models

type DataFile struct {
	Name              string `json:"name"`
	Url               string `json:"url"`
	MimeType      	  string `json:"mimeType"`
	AutoDeleteAt      int64  `json:"autoDeleteAt"`      // in milliseconds
	PrivateUrlExpires uint   `json:"privateUrlExpires"` // in seconds
	UploadedAt        int64  `json:"uploadedAt"`        // in milliseconds
	UpdatedAt         int64  `json:"updatedAt"`         // in milliseconds
	Size              int64  `json:"size"`              // in bytes
	IsPublic          bool   `json:"isPublic"`
}
