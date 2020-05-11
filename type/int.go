package _type

import (
	"database/sql"
)

type Int struct {
	t sql.NullInt64
}

func (i *Int) Value() int {
	return int(i.t.Int64)
}

func (i *Int) IsZero() bool {
	return !i.t.Valid
}

func (i *Int) Scan(value interface{}) error {
	return i.t.Scan(value)
}
