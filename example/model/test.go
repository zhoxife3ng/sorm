package model

import (
	"database/sql"
	"github.com/x554462/go-exception"
	"github.com/x554462/sorm"
)

var TestNotFoundError = exception.New("Test记录未找到", sorm.ModelNotFoundError)

type Test struct {
	sorm.BaseModel
	Id   int            `db:"id,pk"`
	Name sql.NullString `db:"name"`
}

func (t *Test) IndexValues() []interface{} {
	return []interface{}{t.Id}
}

func (t *Test) GetNotFoundError() exception.ErrorWrapper {
	return TestNotFoundError
}
