package service

type Service interface {
	Verify() Verify
}

type ServiceImpl struct {
	VerifySVC Verify
}

func NewServiceImpl(verify Verify) Service {
	return &ServiceImpl{
		VerifySVC: verify,
	}
}

func (si *ServiceImpl) Verify() Verify {
	return si.VerifySVC
}
