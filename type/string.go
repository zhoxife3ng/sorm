package _type

import (
	"database/sql"
	"github.com/x554462/sorm"
)

type String struct {
	t     sql.NullString
	model sorm.Modeller
}

func (s *String) MustValue() string {
	v, err := s.Value()
	if err != nil {
		panic(err)
	}
	return v
}

func (s *String) Value() (string, error) {
	if s.model != nil && !s.model.Loaded() {
		if _, err := s.model.Load(); err != nil {
			return "", err
		}
	}
	return s.t.String, nil
}

func (s *String) IsZero() bool {
	return !s.t.Valid
}

func (s *String) Set(str string) {
	s.t.String = str
	s.t.Valid = true
}

func (s *String) Scan(value interface{}) error {
	return s.t.Scan(value)
}

func (s *String) BindModel(target interface{}) {
	if model, ok := target.(sorm.Modeller); ok {
		s.model = model
	}
}
