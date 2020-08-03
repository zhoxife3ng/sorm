package _type

import (
	"database/sql"
	"github.com/x554462/sorm"
)

type Float struct {
	t     sql.NullFloat64
	model sorm.Modeller
}

func (f *Float) Value() float64 {
	if f.model != nil && !f.model.Loaded() {
		f.model.Load()
	}
	return f.t.Float64
}

func (f *Float) IsZero() bool {
	return !f.t.Valid
}

func (f *Float) Set(ft float64) {
	f.t.Float64 = ft
	f.t.Valid = true
}

func (f *Float) Scan(value interface{}) error {
	return f.t.Scan(value)
}

func (f *Float) BindModel(target interface{}) {
	if model, ok := target.(sorm.Modeller); ok {
		f.model = model
	}
}
