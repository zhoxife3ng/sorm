package sorm

import (
	"reflect"
	"strings"
	"sync"

	"github.com/xkisas/sorm/internal"
)

type ModelIfe interface {
	initBase(dao DaoIfe, indexValues []interface{}, loaded bool)
	CustomDao() DaoIfe
	GetNotFoundError() error
	IndexValues() []interface{}
	GetDaoIfe() DaoIfe
	Loaded() bool
	Load(opts ...Option) (ModelIfe, error)
	Update(set map[string]interface{}) (int64, error)
	Remove() error
	GetId() interface{}
}

type BaseModel struct {
	loaded      bool
	dao         DaoIfe
	indexValues []interface{}
}

func (bm *BaseModel) CustomDao() DaoIfe {
	return &Dao{}
}

func (bm *BaseModel) initBase(dao DaoIfe, indexValues []interface{}, loaded bool) {
	bm.dao = dao
	bm.loaded = loaded
	bm.indexValues = indexValues
}

func (bm *BaseModel) GetDaoIfe() DaoIfe {
	return bm.dao
}

func (bm *BaseModel) Loaded() bool {
	return bm.loaded
}

func (bm *BaseModel) Load(opts ...Option) (ModelIfe, error) {
	option := fetchOption(opts...)
	if option.forUpdate || bm.Loaded() && !option.forceLoad {
		return bm.dao.Select(option.forUpdate, bm.indexValues...)
	}
	where, err := bm.dao.buildWhere(bm.indexValues...)
	if err != nil {
		return nil, err
	}
	return bm.dao.SelectOne(where, opts...)
}

func (bm *BaseModel) Update(set map[string]interface{}) (int64, error) {
	model, err := bm.dao.Select(false, bm.indexValues...)
	if err != nil {
		return 0, err
	}
	var affected int64 = 0
	err = bm.dao.Session().runInTransaction(func() error {
		affected, err = bm.dao.update(model, set)
		return err
	})
	return affected, err
}

func (bm *BaseModel) Remove() error {
	model, err := bm.dao.Select(false, bm.indexValues...)
	if err != nil {
		return err
	}
	return bm.dao.Session().runInTransaction(func() error {
		return bm.dao.remove(model)
	})
}

func (bm *BaseModel) IndexValues() []interface{} {
	return bm.indexValues
}

func (bm *BaseModel) GetNotFoundError() error {
	return ModelNotFoundError
}

func (bm *BaseModel) GetId() interface{} {
	indexValues := bm.IndexValues()
	if len(indexValues) > 0 {
		return indexValues[0]
	}
	return nil
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
