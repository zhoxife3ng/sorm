package builder

import (
	"strings"
	"sync"
)

var (
	builder     sync.Pool
	builderOnce sync.Once
)

func getStrB() *strings.Builder {
	builderOnce.Do(func() {
		builder = sync.Pool{New: func() interface{} {
			return &strings.Builder{}
		}}
	})
	return builder.Get().(*strings.Builder)
}

func putStrB(strB *strings.Builder) {
	strB.Reset()
	builder.Put(strB)
}

type SqlBuilder interface {
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
	var sqlStr = getStrB()
	defer putStrB(sqlStr)
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
	var sqlStr = getStrB()
	defer putStrB(sqlStr)
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
	var sqlStr = getStrB()
	defer putStrB(sqlStr)
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
	var sqlStr = getStrB()
	defer putStrB(sqlStr)
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
	return sqlStr.String(), i.params, nil
}
