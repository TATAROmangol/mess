package jwt

type Service interface {
	GenerateAccessTokenWithSubjectID(userID string) (string, error)
	GenerateRefreshTokenWithSubjectID(userID string) (string, error)
	GetSubjectIDFromToken(token string) (string, error)
}
