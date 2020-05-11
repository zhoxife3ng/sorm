package model

import (
	"github.com/x554462/go-exception"
	"github.com/x554462/sorm"
	"github.com/x554462/sorm/type"
)

var TestNotFoundError = exception.New("Test记录未找到", sorm.ModelNotFoundError)

type Test struct {
	sorm.BaseModel
	Id   int          `db:"id,pk"`
	Name _type.String `db:"name"`
}

func (t *Test) IndexValues() []interface{} {
	return []interface{}{t.Id}
}

func (t *Test) GetNotFoundError() exception.ErrorWrapper {
	return TestNotFoundError
}
