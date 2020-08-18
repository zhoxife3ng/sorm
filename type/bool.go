package _type

import (
	"database/sql"
	"github.com/x554462/sorm"
)

type Bool struct {
	t     sql.NullBool
	model sorm.ModelIfe
}

func (b *Bool) MustValue() bool {
	v, err := b.Value()
	if err != nil {
		panic(err)
	}
	return v
}

func (b *Bool) Value() (bool, error) {
	if b.model != nil && !b.model.Loaded() {
		if _, err := b.model.Load(); err != nil {
			return false, err
		}
	}
	return b.t.Bool, nil
}

func (b *Bool) Set(bl bool) {
	b.t.Bool = bl
	b.t.Valid = true
}

func (b *Bool) MustIsZero() bool {
	b.MustValue()
	return !b.t.Valid
}

func (b *Bool) IsZero() (bool, error) {
	if _, err := b.Value(); err != nil {
		return false, err
	}
	return !b.t.Valid, nil
}

func (b *Bool) Scan(value interface{}) error {
	return b.t.Scan(value)
}

func (b *Bool) BindModel(target interface{}) {
	if model, ok := target.(sorm.ModelIfe); ok {
		b.model = model
	}
}
