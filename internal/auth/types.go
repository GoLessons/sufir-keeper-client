package auth

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int
}

type UserInfo struct {
	UserID string
}

