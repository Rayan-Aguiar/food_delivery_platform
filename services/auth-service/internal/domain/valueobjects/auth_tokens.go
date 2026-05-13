package valueobjects

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int64
}


