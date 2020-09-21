package _type

import (
	"database/sql"
	"github.com/xkisas/sorm"
	"github.com/xkisas/sorm/internal"
)

type Map struct {
	loaded bool
	t      sql.NullString
	model  sorm.ModelIfe
}

func (m *Map) MustValue() map[string]interface{} {
	v, err := m.Value()
	if err != nil {
		panic(err)
	}
	return v
}

func (m *Map) Value() (map[string]interface{}, error) {
	if !m.loaded && m.model != nil && !m.model.Loaded() {
		if _, err := m.model.Load(); err != nil {
			return nil, err
		}
	}
	if !m.t.Valid {
		return nil, nil
	}
	var data map[string]interface{}
	err := internal.JsonUnmarshal(internal.StringToBytes(m.t.String), &data)
	return data, err
}

func (m *Map) MustIsZero() bool {
	m.MustValue()
	return !m.t.Valid
}

func (m *Map) IsZero() (bool, error) {
	if _, err := m.Value(); err != nil {
		return false, err
	}
	return !m.t.Valid, nil
}

func (m *Map) Set(mp map[string]interface{}) {
	b, _ := internal.JsonMarshal(mp)
	m.t.String = internal.BytesToString(b)
	m.t.Valid = true
	m.loaded = true
}

func (m *Map) Scan(value interface{}) error {
	m.loaded = true
	return m.t.Scan(value)
}

func (m *Map) BindModel(target interface{}) {
	if model, ok := target.(sorm.ModelIfe); ok {
		m.model = model
	}
}
