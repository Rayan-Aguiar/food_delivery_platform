package valueobjects

import "testing"

func TestAuthTokensShape(t *testing.T) {
	tokens := AuthTokens{
		AccessToken:  "access",
		RefreshToken: "refresh",
		TokenType:    "Bearer",
		ExpiresIn:    900,
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" || tokens.TokenType != "Bearer" || tokens.ExpiresIn <= 0 {
		t.Fatalf("unexpected auth tokens: %+v", tokens)
	}
}
