package sorm

import (
	"github.com/x554462/go-exception"
	"github.com/x554462/sorm/internal"
	"reflect"
	"strings"
	"sync"
)

type Modeller interface {
	initBase(dao *Dao, m Modeller, loaded bool)
	GetNotFoundError() exception.ErrorWrapper
	IndexValues() []interface{}
	GetDao() *Dao
	Loaded() bool
	Load(forUpdate bool, force ...bool) Modeller
	Update(set map[string]interface{}) int64
	Remove()
}

type BaseModel struct {
	loaded      bool
	dao         *Dao
	indexValues []interface{}
}

func (bm *BaseModel) initBase(dao *Dao, m Modeller, loaded bool) {
	bm.dao = dao
	bm.indexValues = m.IndexValues()
	bm.loaded = loaded
}

func (bm *BaseModel) GetDao() *Dao {
	return bm.dao
}

func (bm *BaseModel) Loaded() bool {
	return bm.loaded
}

func (bm *BaseModel) Load(forUpdate bool, force ...bool) Modeller {
	if !forUpdate && (bm.loaded == false || len(force) > 0 && force[0]) {
		// 主键查询直接走主库
		return bm.dao.SelectOne(bm.dao.buildWhere(bm.indexValues...), false)
	}
	return bm.dao.Select(forUpdate, bm.indexValues...)
}

func (bm *BaseModel) Update(set map[string]interface{}) int64 {
	return bm.dao.update(bm.dao.Select(false, bm.indexValues...), set)
}

func (bm *BaseModel) Remove() {
	bm.dao.remove(bm.dao.Select(false, bm.indexValues...))
}

func (bm *BaseModel) GetNotFoundError() exception.ErrorWrapper {
	return ModelNotFoundError
}

// table info
// 自动解析Model
type tableInfo struct {
	tableName   string
	indexFields []string
}

var tableInfos = sync.Map{}

func parseTableInfo(modelType reflect.Type) (string, []string) {
	name := modelType.Name()
	if v, ok := tableInfos.Load(name); ok {
		tableInfo := v.(tableInfo)
		return tableInfo.tableName, tableInfo.indexFields
	}
	var indexFields = make([]string, 0)
	for i := 0; i < modelType.NumField(); i++ {
		fieldType := modelType.Field(i)
		if tag, ok := fieldType.Tag.Lookup(defaultTagName); ok {
			idx := strings.IndexByte(tag, ',')
			if -1 != idx {
				if tag[:idx] == "pk" {
					indexFields = append(indexFields, tag[idx+1:])
				} else if tag[idx+1:] == "pk" {
					indexFields = append(indexFields, tag[:idx])
				}
			}
		}
	}
	tableInfo := tableInfo{
		tableName:   internal.TitleSnakeName(name),
		indexFields: indexFields,
	}
	tableInfos.Store(name, tableInfo)
	return tableInfo.tableName, tableInfo.indexFields
}
