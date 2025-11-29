package service

type Service interface {
	TokenService() Token
	VerifyService() Verify
}

type ServiceImpl struct {
	Token
	Verify
}

func NewServiceImpl(token Token, verify Verify) *ServiceImpl {
	return &ServiceImpl{
		Token:  token,
		Verify: verify,
	}
}
