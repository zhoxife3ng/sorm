package _type

import (
	"database/sql"
	"github.com/x554462/sorm"
	"time"
)

type Time struct {
	t     sql.NullTime
	model sorm.Modeller
}

func (t *Time) MustValue() time.Time {
	v, err := t.Value()
	if err != nil {
		panic(err)
	}
	return v
}

func (t *Time) Value() (time.Time, error) {
	if t.model != nil && !t.model.Loaded() {
		if _, err := t.model.Load(); err != nil {
			return time.Time{}, err
		}
	}
	return t.t.Time, nil
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

func (t *Time) BindModel(target interface{}) {
	if model, ok := target.(sorm.Modeller); ok {
		t.model = model
	}
}
