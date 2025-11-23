package service

type Service interface {
	TokenService() Token
	VerifyService() Verify
}

type ServiceImpl struct {
	Token
	Verify
}

func NewServiceImpl(tokenSvc Token, verifySvc Verify) *ServiceImpl {
	return &ServiceImpl{
		Token:  tokenSvc,
		Verify: verifySvc,
	}
}
