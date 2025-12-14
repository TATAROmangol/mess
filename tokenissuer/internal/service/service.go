package service

type Service interface {
	VerifyService() Verify
}

type ServiceImpl struct {
	Verify
}

func NewServiceImpl(verify Verify) *ServiceImpl {
	return &ServiceImpl{
		Verify: verify,
	}
}
