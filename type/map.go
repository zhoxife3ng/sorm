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

func (m *Map) Value() map[string]interface{} {
	if m.model != nil && !m.model.Loaded() {
		m.model.Load()
	}
	if m.IsZero() {
		return nil
	}
	var data map[string]interface{}
	internal.JsonUnmarshal(internal.StringToBytes(m.t.String), &data)
	return data
}

func (m *Map) IsZero() bool {
	return !m.t.Valid
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
