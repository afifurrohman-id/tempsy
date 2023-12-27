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

type GuestToken struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int    `json:"expiresIn"` // in seconds
}

// GOAuth2Token
// For unmarshal purpose
type GOAuth2Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scopes       string `json:"scope"` // space-separated list of scopes
	IdToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // in seconds
}
