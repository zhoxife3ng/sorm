package _type

import (
	"database/sql"
	"github.com/xkisas/sorm"
)

type Int struct {
	loaded bool
	t      sql.NullInt64
	model  sorm.ModelIfe
}

func (i *Int) MustValue() int {
	v, err := i.Value()
	if err != nil {
		panic(err)
	}
	return v
}

func (i *Int) Value() (int, error) {
	if !i.loaded && i.model != nil && !i.model.Loaded() {
		if _, err := i.model.Load(); err != nil {
			return 0, err
		}
	}
	return int(i.t.Int64), nil
}

func (i *Int) MustIsZero() bool {
	i.MustValue()
	return !i.t.Valid
}

func (i *Int) IsZero() (bool, error) {
	if _, err := i.Value(); err != nil {
		return false, err
	}
	return !i.t.Valid, nil
}

func (i *Int) Set(it int) {
	i.t.Int64 = int64(it)
	i.t.Valid = true
	i.loaded = true
}

func (i *Int) Scan(value interface{}) error {
	i.loaded = true
	return i.t.Scan(value)
}

func (i *Int) BindModel(target interface{}) {
	if model, ok := target.(sorm.ModelIfe); ok {
		i.model = model
	}
}
