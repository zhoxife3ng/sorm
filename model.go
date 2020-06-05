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
	Load(opts ...option) (Modeller, error)
	Update(set map[string]interface{}) (int64, error)
	Remove() error
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

func (bm *BaseModel) Load(opts ...option) (Modeller, error) {
	options := newOptions()
	for _, o := range opts {
		o(&options)
	}
	if options.forUpdate || bm.Loaded() && !options.forceLoad {
		return bm.dao.Select(options.forUpdate, bm.indexValues...)
	}
	return bm.dao.SelectOne(bm.dao.buildWhere(bm.indexValues...), opts...)
}

func (bm *BaseModel) Update(set map[string]interface{}) (int64, error) {
	model, err := bm.dao.Select(false, bm.indexValues...)
	if err != nil {
		return 0, err
	}
	return bm.dao.update(model, set)
}

func (bm *BaseModel) Remove() error {
	model, err := bm.dao.Select(false, bm.indexValues...)
	if err != nil {
		return err
	}
	return bm.dao.remove(model)
}

func (bm *BaseModel) GetNotFoundError() exception.ErrorWrapper {
	return ModelNotFoundError
}

// table info
// 自动解析Model
type tableInfo struct {
	tableName   string
	indexFields []string
	fields      []string
}

var tableInfos = sync.Map{}

func parseTableInfo(modelType reflect.Type) (string, []string, []string) {
	name := modelType.Name()
	if v, ok := tableInfos.Load(name); ok {
		tableInfo := v.(tableInfo)
		return tableInfo.tableName, tableInfo.indexFields, tableInfo.fields
	}
	var indexFields, fields = make([]string, 0), make([]string, 0)
	for i := 0; i < modelType.NumField(); i++ {
		fieldType := modelType.Field(i)
		if tag, ok := fieldType.Tag.Lookup(defaultTagName); ok {
			idx := strings.IndexByte(tag, ',')
			if -1 != idx {
				if tag[:idx] == "pk" {
					indexFields = append(indexFields, tag[idx+1:])
					fields = append(fields, tag[idx+1:])
				} else if tag[idx+1:] == "pk" {
					indexFields = append(indexFields, tag[:idx])
					fields = append(fields, tag[:idx])
				}
			} else {
				fields = append(fields, tag)
			}
		}
	}
	tableInfo := tableInfo{
		tableName:   internal.TitleSnakeName(name),
		indexFields: indexFields,
		fields:      fields,
	}
	tableInfos.Store(name, tableInfo)
	return tableInfo.tableName, tableInfo.indexFields, tableInfo.fields
}
