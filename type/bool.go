package _type

import (
	"database/sql"
	"github.com/x554462/sorm"
)

type Bool struct {
	t     sql.NullBool
	model sorm.Modeller
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

func (b *Bool) IsZero() bool {
	return !b.t.Valid
}

func (b *Bool) Scan(value interface{}) error {
	return b.t.Scan(value)
}

func (b *Bool) BindModel(target interface{}) {
	if model, ok := target.(sorm.Modeller); ok {
		b.model = model
	}
}
