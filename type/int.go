package _type

import (
	"database/sql"
	"github.com/x554462/sorm"
)

type Int struct {
	t     sql.NullInt64
	model sorm.Modeller
}

func (i *Int) Value() int {
	if i.model != nil && !i.model.Loaded() {
		i.model.Load()
	}
	return int(i.t.Int64)
}

func (i *Int) IsZero() bool {
	return !i.t.Valid
}

func (i *Int) Set(it int) {
	i.t.Int64 = int64(it)
	i.t.Valid = true
}

func (i *Int) Scan(value interface{}) error {
	return i.t.Scan(value)
}

func (i *Int) BindModel(target interface{}) {
	if model, ok := target.(sorm.Modeller); ok {
		i.model = model
	}
}
