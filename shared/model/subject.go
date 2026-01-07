package model

type Subject interface {
	GetSubjectId() string
	GetEmail() string
}

type SubjectIMPL struct {
	SubjectID string
	Name      string
	Email     string
}

func (s *SubjectIMPL) GetSubjectId() string {
	return s.SubjectID
}

func (s *SubjectIMPL) GetEmail() string {
	return s.Email
}
