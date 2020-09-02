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

var selectSeq = []func(*Selector) (string, error){
	(*Selector).processSelect,
	(*Selector).processForceIndex,
	(*Selector).processJoins,
	(*Selector).processWhere,
	(*Selector).processGroup,
	(*Selector).processHaving,
	(*Selector).processOrder,
	(*Selector).processLimit,
	(*Selector).processOffset,
	(*Selector).processTail,
}

var updateSeq = []func(*Updater) (string, error){
	(*Updater).processUpdate,
	(*Updater).processJoins,
	(*Updater).processSet,
	(*Updater).processWhere,
}

var deleteSeq = []func(*Deleter) (string, error){
	(*Deleter).processDelete,
	(*Deleter).processWhere,
}

var insertSeq = []func(*Inserter) (string, error){
	(*Inserter).processInsert,
}

func (s *Selector) Build() (string, []interface{}, error) {
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

func (u *Updater) Build() (string, []interface{}, error) {
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
func (d *Deleter) Build() (string, []interface{}, error) {
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
func (i *Inserter) Build() (string, []interface{}, error) {
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
