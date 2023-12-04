package models

type User struct {
	UserName   string `json:"username"`
	TotalFiles int    `json:"totalFiles"`
}

// GoogleAccountInfo
// For unmarshal purpose
type GoogleAccountInfo struct {
	*User
	Email         string `json:"email"`
	Picture       string `json:"picture"`
	ID            string `json:"id"`
	VerifiedEmail bool   `json:"verified_email"`
}

type Token struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int    `json:"expiresIn"` // in seconds
	TokenType   string `json:"tokenType"`
}

// GOAuth2Token
// For unmarshal purpose
type GOAuth2Token struct {
	*Token
	Scopes       string `json:"scope"` // separated by space
	IdToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
}
