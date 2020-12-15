package _type

import (
	"database/sql"

	"github.com/xkisas/sorm"
)

type Float struct {
	loaded bool
	t      sql.NullFloat64
	model  sorm.ModelIfe
}

func (f *Float) MustValue() float64 {
	v, err := f.Value()
	if err != nil {
		panic(err)
	}
	return v
}

func (f *Float) Value() (float64, error) {
	if !f.loaded && f.model != nil && !f.model.Loaded() {
		if _, err := f.model.Load(); err != nil {
			return 0, err
		}
	}
	return f.t.Float64, nil
}

func (f *Float) MustIsNull() bool {
	f.MustValue()
	return !f.t.Valid
}

func (f *Float) IsNull() (bool, error) {
	if _, err := f.Value(); err != nil {
		return false, err
	}
	return !f.t.Valid, nil
}

func (f *Float) Set(ft float64) {
	f.t.Float64 = ft
	f.t.Valid = true
	f.loaded = true
}

func (f *Float) Scan(value interface{}) error {
	f.loaded = true
	return f.t.Scan(value)
}

func (f *Float) BindModel(target interface{}) {
	if model, ok := target.(sorm.ModelIfe); ok {
		f.model = model
	}
}
