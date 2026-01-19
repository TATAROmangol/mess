package domain

import (
	"fmt"

	"github.com/TATAROmangol/mess/chat/internal/storage"
)

var (
	SubjectNotHaveThisResource = fmt.Errorf("subject not have this resource")
	ErrNotFound                = storage.ErrNoRows
)
