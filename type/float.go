package _type

import (
	"database/sql"
	"github.com/x554462/sorm"
)

type Float struct {
	t     sql.NullFloat64
	model sorm.Modeller
}

func (f *Float) MustValue() float64 {
	v, err := f.Value()
	if err != nil {
		panic(err)
	}
	return v
}

func (f *Float) Value() (float64, error) {
	if f.model != nil && !f.model.Loaded() {
		if _, err := f.model.Load(); err != nil {
			return 0, err
		}
	}
	return f.t.Float64, nil
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
