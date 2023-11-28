package internal

type OAuth2Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type User struct {
	UserName   string `json:"username"`
	TotalFiles int    `json:"total_files"`
	Picture    string `json:"picture"`
}

type GoogleAccountInfo struct {
	*User
	Email         string `json:"email"`
	ID            string `json:"id"`
	VerifiedEmail bool   `json:"verified_email"`
}

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

type Token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"` // in seconds
	TokenType   string `json:"token_type"`
}

type GOAuth2Token struct {
	*Token
	Scopes       string `json:"scope"` // separated by space
	IdToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
}

type GOAuth2Config struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	CallbackURL  string   `json:"callback_url"`
	Scopes       []string `json:"scopes"`
}

type ApiError struct {
	Type        string `json:"error_type"`
	Description string `json:"error_description"`
}
