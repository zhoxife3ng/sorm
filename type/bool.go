package _type

import (
	"database/sql"
	"github.com/x554462/sorm"
)

type Bool struct {
	t     sql.NullBool
	model sorm.Modeller
}

func (b *Bool) Value() bool {
	if b.model != nil && !b.model.Loaded() {
		b.model.Load()
	}
	return b.t.Bool
}

func (b *Bool) IsZero() bool {
	return !b.t.Valid
}

func (b *Bool) Set(bl bool) {
	b.t.Bool = bl
	b.t.Valid = true
}

func (b *Bool) Scan(value interface{}) error {
	return b.t.Scan(value)
}

func (b *Bool) BindModel(target interface{}) {
	if model, ok := target.(sorm.Modeller); ok {
		b.model = model
	}
}
