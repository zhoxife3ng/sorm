package model

import (
	"github.com/x554462/sorm"
	"github.com/x554462/sorm/type"
)

var TestNotFoundError = sorm.NewError(sorm.ModelNotFoundError, "Test记录未找到")

type Test struct {
	sorm.BaseModel
	Id   int          `db:"id,pk"`
	Name _type.String `db:"name"`
	Time _type.Time   `db:"time"`
}

func (t *Test) GetNotFoundError() error {
	return TestNotFoundError
}
