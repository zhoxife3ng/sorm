package _type

import (
	"database/sql"
	"time"

	"github.com/xkisas/sorm"
)

type Time struct {
	loaded bool
	t      sql.NullTime
	model  sorm.ModelIfe
}

func (t *Time) MustValue() time.Time {
	v, err := t.Value()
	if err != nil {
		panic(err)
	}
	return v
}

func (t *Time) Value() (time.Time, error) {
	if !t.loaded && t.model != nil && !t.model.Loaded() {
		if _, err := t.model.Load(); err != nil {
			return time.Time{}, err
		}
	}
	return t.t.Time, nil
}

func (t *Time) MustIsNull() bool {
	t.MustValue()
	return !t.t.Valid
}

func (t *Time) IsNull() (bool, error) {
	if _, err := t.Value(); err != nil {
		return false, err
	}
	return !t.t.Valid, nil
}

func (t *Time) Set(tm time.Time) {
	t.t.Time = tm
	t.t.Valid = true
	t.loaded = true
}

func (t *Time) Scan(value interface{}) error {
	t.loaded = true
	return t.t.Scan(value)
}

func (t *Time) BindModel(target interface{}) {
	if model, ok := target.(sorm.ModelIfe); ok {
		t.model = model
	}
}
