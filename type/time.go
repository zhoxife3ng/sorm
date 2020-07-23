package _type

import (
	"database/sql"
	"time"
)

type Time struct {
	t sql.NullTime
}

func (t *Time) Value() time.Time {
	return t.t.Time
}

func (t *Time) IsZero() bool {
	return !t.t.Valid
}

func (t *Time) Set(tm time.Time) {
	t.t.Time = tm
	t.t.Valid = true
}

func (t *Time) Scan(value interface{}) error {
	return t.t.Scan(value)
}
