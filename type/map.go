package _type

import (
	"database/sql"
	"github.com/x554462/sorm"
	"github.com/x554462/sorm/internal"
)

type Map struct {
	t     sql.NullString
	model sorm.Modeller
}

func (m *Map) MustValue() map[string]interface{} {
	v, err := m.Value()
	if err != nil {
		panic(err)
	}
	return v
}

func (m *Map) Value() (map[string]interface{}, error) {
	if m.model != nil && !m.model.Loaded() {
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
}

func (m *Map) Scan(value interface{}) error {
	return m.t.Scan(value)
}

func (m *Map) BindModel(target interface{}) {
	if model, ok := target.(sorm.Modeller); ok {
		m.model = model
	}
}
