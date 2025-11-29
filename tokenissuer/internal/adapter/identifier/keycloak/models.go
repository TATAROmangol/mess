package keycloak

type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
}

func (t *TokenResponse) GetAccessToken() string {
	return t.AccessToken
}

func (t *TokenResponse) GetRefreshToken() string {
	return t.RefreshToken
}

func (t *TokenResponse) GetExpiresIn() int {
	return t.ExpiresIn
}

func (t *TokenResponse) GetRefreshExpiresIn() int {
	return t.RefreshExpiresIn
}

func (t *TokenResponse) GetTokenType() string {
	return t.TokenType
}
