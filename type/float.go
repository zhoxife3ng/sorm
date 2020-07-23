package _type

import "database/sql"

type Float struct {
	t sql.NullFloat64
}

func (f *Float) Value() float64 {
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
