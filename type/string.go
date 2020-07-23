package _type

import "database/sql"

type String struct {
	t sql.NullString
}

func (s *String) Value() string {
	return s.t.String
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
