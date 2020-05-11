package _type

import "database/sql"

type Bool struct {
	t sql.NullBool
}

func (b *Bool) Value() bool {
	return b.t.Bool
}

func (b *Bool) IsZero() bool {
	return !b.t.Valid
}

func (b *Bool) Scan(value interface{}) error {
	return b.t.Scan(value)
}
