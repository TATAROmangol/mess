package domain

type Service interface {
	TokenService() TokenService
	VerifyService() VerifyService
}

type Domain struct {
	TokenService
	VerifyService
}

func NewDomain(tokenSvc TokenService, verifySvc VerifyService) *Domain {
	return &Domain{
		TokenService:  tokenSvc,
		VerifyService: verifySvc,
	}
}
