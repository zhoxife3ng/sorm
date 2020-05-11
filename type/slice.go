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

func (s *Slice) Scan(value interface{}) error {
	return s.t.Scan(value)
}
