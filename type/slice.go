package _type

import (
	"database/sql"
	"github.com/x554462/sorm/internal"
)

type Slice struct {
	t sql.NullString
}

func (s *Slice) Value() []interface{} {
	if s.IsZero() {
		return nil
	}
	var data []interface{}
	internal.JsonUnmarshal(internal.StringToBytes(s.t.String), &data)
	return data
}

func (s *Slice) IsZero() bool {
	return !s.t.Valid
}

func (s *Slice) Set(sl map[string]interface{}) {
	b, _ := internal.JsonMarshal(sl)
	s.t.String = internal.BytesToString(b)
	s.t.Valid = true
}

func (s *Slice) Scan(value interface{}) error {
	return s.t.Scan(value)
}
