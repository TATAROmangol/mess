package auth

import "github.com/TATAROmangol/mess/shared/model"

type Service interface {
	Verify(src string) (model.Subject, error)
}
