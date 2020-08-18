package builder

import (
	"errors"
	"strings"
	"sync"
)

var (
	ErrBuildPlaceHolder  = errors.New("[builder] place holder num does not match with values num")
	ErrNotSupportProcess = errors.New("[builder] not support process")
	ErrProcessOrder      = errors.New("[builder] process order error")
	ErrProcessSet        = errors.New("[builder] process set error")
)

var (
	builder     sync.Pool
	builderOnce sync.Once
)

func getStrBuilder() *strings.Builder {
	builderOnce.Do(func() {
		builder = sync.Pool{New: func() interface{} {
			return &strings.Builder{}
		}}
	})
	return builder.Get().(*strings.Builder)
}

func putStrBuilder(str *strings.Builder) {
	str.Reset()
	builder.Put(str)
}

type SqlBuildIfe interface {
	Build() (string, []interface{}, error)
}

var selectSeq = []func(*baseSelect) (string, error){
	(*baseSelect).processSelect,
	(*baseSelect).processForceIndex,
	(*baseSelect).processJoins,
	(*baseSelect).processWhere,
	(*baseSelect).processGroup,
	(*baseSelect).processHaving,
	(*baseSelect).processOrder,
	(*baseSelect).processLimit,
	(*baseSelect).processOffset,
	(*baseSelect).processTail,
}

var updateSeq = []func(*baseUpdate) (string, error){
	(*baseUpdate).processUpdate,
	(*baseUpdate).processJoins,
	(*baseUpdate).processSet,
	(*baseUpdate).processWhere,
}

var deleteSeq = []func(*baseDelete) (string, error){
	(*baseDelete).processDelete,
	(*baseDelete).processWhere,
}

var insertSeq = []func(*baseInsert) (string, error){
	(*baseInsert).processInsert,
}

func (s *baseSelect) Build() (string, []interface{}, error) {
	var sqlStr = getStrBuilder()
	defer putStrBuilder(sqlStr)
	for _, m := range selectSeq {
		sql, err := m(s)
		if err != nil {
			return "", nil, err
		} else if sql != "" {
			if sqlStr.Len() > 0 {
				sqlStr.WriteString(" ")
			}
			sqlStr.WriteString(sql)
		}
	}
	return sqlStr.String(), s.params, nil
}

func (u *baseUpdate) Build() (string, []interface{}, error) {
	var sqlStr = getStrBuilder()
	defer putStrBuilder(sqlStr)
	for _, m := range updateSeq {
		sql, err := m(u)
		if err != nil {
			return "", nil, err
		} else if sql != "" {
			if sqlStr.Len() > 0 {
				sqlStr.WriteString(" ")
			}
			sqlStr.WriteString(sql)
		}
	}
	return sqlStr.String(), u.params, nil
}
func (d *baseDelete) Build() (string, []interface{}, error) {
	var sqlStr = getStrBuilder()
	defer putStrBuilder(sqlStr)
	for _, m := range deleteSeq {
		sql, err := m(d)
		if err != nil {
			return "", nil, err
		} else if sql != "" {
			if sqlStr.Len() > 0 {
				sqlStr.WriteString(" ")
			}
			sqlStr.WriteString(sql)
		}
	}
	return sqlStr.String(), d.params, nil
}
func (i *baseInsert) Build() (string, []interface{}, error) {
	var sqlStr = getStrBuilder()
	defer putStrBuilder(sqlStr)
	for _, m := range insertSeq {
		sql, err := m(i)
		if err != nil {
			return "", nil, err
		} else if sql != "" {
			if sqlStr.Len() > 0 {
				sqlStr.WriteString(" ")
			}
			sqlStr.WriteString(sql)
		}
	}
	var params = make([]interface{}, 0)
	for _, p := range i.params {
		params = append(params, p...)
		for j := len(p); j < len(i.columns); j++ {
			params = append(params, nil)
		}
	}
	return sqlStr.String(), params, nil
}
