package _type

import (
	"database/sql"

	"github.com/xkisas/sorm"
)

type Bool struct {
	loaded bool
	t      sql.NullBool
	model  sorm.ModelIfe
}

func (b *Bool) MustValue() bool {
	v, err := b.Value()
	if err != nil {
		panic(err)
	}
	return v
}

func (b *Bool) Value() (bool, error) {
	if !b.loaded && b.model != nil && !b.model.Loaded() {
		if _, err := b.model.Load(); err != nil {
			return false, err
		}
	}
	return b.t.Bool, nil
}

func (b *Bool) MustIsNull() bool {
	b.MustValue()
	return !b.t.Valid
}

func (b *Bool) IsNull() (bool, error) {
	if _, err := b.Value(); err != nil {
		return false, err
	}
	return !b.t.Valid, nil
}

func (b *Bool) Set(bl bool) {
	b.t.Bool = bl
	b.t.Valid = true
	b.loaded = true
}

func (b *Bool) Scan(value interface{}) error {
	b.loaded = true
	return b.t.Scan(value)
}

func (b *Bool) BindModel(target interface{}) {
	if model, ok := target.(sorm.ModelIfe); ok {
		b.model = model
	}
}
