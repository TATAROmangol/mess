package jwt

type Service interface {
	GenerateAccessTokenWithUserID(userID string) (string, error)
	GenerateRefreshTokenWithUserID(userID string) (string, error)
	GetUserIDFromToken(token string) (string, error)
}
