package _type

import (
	"database/sql"
	"github.com/xkisas/sorm"
	"github.com/xkisas/sorm/internal"
)

type Slice struct {
	t     sql.NullString
	model sorm.ModelIfe
}

func (s *Slice) MustValue() []interface{} {
	v, err := s.Value()
	if err != nil {
		panic(err)
	}
	return v
}

func (s *Slice) Value() ([]interface{}, error) {
	if s.model != nil && !s.model.Loaded() {
		if _, err := s.model.Load(); err != nil {
			return nil, err
		}
	}
	if !s.t.Valid {
		return nil, nil
	}
	var data []interface{}
	err := internal.JsonUnmarshal(internal.StringToBytes(s.t.String), &data)
	return data, err
}

func (s *Slice) MustIsZero() bool {
	s.MustValue()
	return !s.t.Valid
}

func (s *Slice) IsZero() (bool, error) {
	if _, err := s.Value(); err != nil {
		return false, err
	}
	return !s.t.Valid, nil
}

func (s *Slice) Set(sl map[string]interface{}) {
	b, _ := internal.JsonMarshal(sl)
	s.t.String = internal.BytesToString(b)
	s.t.Valid = true
}

func (s *Slice) Scan(value interface{}) error {
	return s.t.Scan(value)
}

func (s *Slice) BindModel(target interface{}) {
	if model, ok := target.(sorm.ModelIfe); ok {
		s.model = model
	}
}
