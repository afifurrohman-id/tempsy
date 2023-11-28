package models

type User struct {
	UserName   string `json:"username"`
	TotalFiles int    `json:"total_files"`
}

type GoogleAccountInfo struct {
	*User
	Email         string `json:"email"`
	Picture       string `json:"picture"`
	ID            string `json:"id"`
	VerifiedEmail bool   `json:"verified_email"`
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
