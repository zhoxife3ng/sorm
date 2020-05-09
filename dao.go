package sorm

import (
	"database/sql"
	"github.com/didi/gendry/builder"
	"github.com/x554462/go-exception"
	"github.com/x554462/sorm/db"
	"github.com/x554462/sorm/util"
	"reflect"
)

const defaultTagName = "db"

type Dao struct {
	// dao绑定的表
	tableName string
	// 主键字段
	indexFields []string
	// 记录未找到时报错
	notFoundError exception.ErrorWrapper
	// 绑定session
	daoSession *Session
	// 空model，通过反射可用于构造model对象
	modelType reflect.Type
}

var (
	ModelRuntimeError  = exception.New("model runtime error", exception.RootError)
	ModelNotFoundError = exception.New("model not found error", exception.RootError)
)

func (d *Dao) GetDaoSession() *Session {
	return d.daoSession
}

func (d *Dao) CheckError(err error) {
	if err != nil {
		exception.ThrowMsgWithCallerDepth(err.Error(), ModelRuntimeError, 3)
	}
}

func (d *Dao) newModel(data map[string]interface{}) Modeller {
	if indexValues, ok := d.getIndexValuesFromData(data); ok {
		if model := d.query(indexValues...); model != nil {
			return model
		}
	}
	vc := reflect.New(d.modelType)
	model, ok := vc.Interface().(Modeller)
	if !ok {
		exception.ThrowMsg("dao.newModel error", ModelRuntimeError)
	}
	return model
}

func (d *Dao) afterScanModel(model Modeller, loaded bool) {
	model.initBase(d, model, loaded)
	d.save(model)
}

func (d *Dao) buildWhere(indexes ...interface{}) map[string]interface{} {
	if len(d.indexFields) != len(indexes) {
		exception.ThrowMsg("dao.buildWhere index number error", ModelRuntimeError)
	}
	where := make(map[string]interface{})
	for i, v := range d.indexFields {
		where[v] = indexes[i]
	}
	return where
}

func (d *Dao) getIndexValuesFromData(data map[string]interface{}) ([]interface{}, bool) {
	indexValues := make([]interface{}, 0)
	for _, v := range d.indexFields {
		if iv, ok := data[v]; ok {
			indexValues = append(indexValues, iv)
		} else {
			return nil, false
		}
	}
	return indexValues, true
}

func (d *Dao) createOneFromRows(rows *sql.Rows) Modeller {
	defer rows.Close()
	m := d.ResolveDataFromRows(rows)
	if len(m) < 1 {
		d.notFoundError.Throw()
	}
	return d.createOne(m[0], true)
}

// 创建单个model对象
func (d *Dao) createOne(data map[string]interface{}, loaded bool) Modeller {
	model := d.newModel(data)
	d.CheckError(util.ScanStruct(data, model, defaultTagName))
	d.afterScanModel(model, loaded)
	return model
}

func (d *Dao) createMultiFromRows(rows *sql.Rows) []Modeller {
	defer rows.Close()
	return d.createMulti(d.ResolveDataFromRows(rows))
}

// 创建多个model对象
func (d *Dao) createMulti(data []map[string]interface{}) []Modeller {
	modelIs := make([]Modeller, 0)
	for _, v := range data {
		model := d.newModel(v)
		d.CheckError(util.ScanStruct(v, model, defaultTagName))
		modelIs = append(modelIs, model)
		d.afterScanModel(model, true)
	}
	return modelIs
}

func (d *Dao) GetTableName() string {
	return d.tableName
}

func (d *Dao) Select(forUpdate bool, indexes ...interface{}) Modeller {
	if forUpdate {
		daoSession := d.GetDaoSession()
		cond, vals, err := builder.BuildSelect(d.GetTableName(), d.buildWhere(indexes...), nil)
		d.CheckError(err)
		if daoSession.tx == nil {
			exception.ThrowMsg("Attempt to load for update out of transaction", ModelRuntimeError)
		}
		cond = cond + " FOR UPDATE"
		row, err := daoSession.Query(cond, vals...)
		d.CheckError(err)
		return d.createOneFromRows(row)
	}
	obj := d.query(indexes...)
	if obj != nil {
		return obj
	}
	return d.createOne(d.buildWhere(indexes...), false)
}

func (d *Dao) Insert(data map[string]interface{}, indexes ...interface{}) Modeller {
	cond, vals, err := builder.BuildInsert(d.GetTableName(), []map[string]interface{}{data})
	d.CheckError(err)
	result, err := d.GetDaoSession().Exec(cond, vals...)
	d.CheckError(err)
	if affected, _ := result.RowsAffected(); affected != 1 {
		exception.ThrowMsg("dao.Insert error", ModelRuntimeError)
	}
	if len(indexes) > 0 {
		for i, index := range indexes {
			data[d.indexFields[i]] = index
		}
	} else if len(d.indexFields) == 1 {
		if id, err := result.LastInsertId(); err == nil {
			data[d.indexFields[0]] = id
		}
	}
	var m = d.newModel(data)
	d.CheckError(util.ScanStruct(data, m, defaultTagName))
	d.afterScanModel(m, true)
	return m
}

func (d *Dao) Update(model Modeller, data map[string]interface{}) int64 {
	cond, vals, err := builder.BuildUpdate(d.GetTableName(), d.buildWhere(model.IndexValues()...), data)
	d.CheckError(err)
	result, err := d.GetDaoSession().Exec(cond, vals...)
	d.CheckError(err)
	affected, _ := result.RowsAffected()
	if affected == 1 {
		util.ScanStruct(data, model, defaultTagName)
		d.afterScanModel(model, true)
	}
	return affected
}

func (d *Dao) Remove(model Modeller) {
	cond, vals, err := builder.BuildDelete(d.GetTableName(), d.buildWhere(model.IndexValues()...))
	d.CheckError(err)
	_, err = d.GetDaoSession().Exec(cond, vals...)
	d.CheckError(err)
}

func (d *Dao) SelectOne(where map[string]interface{}, useSlave ...bool) Modeller {
	cond, vals, err := builder.BuildSelect(d.GetTableName(), where, nil)
	d.CheckError(err)
	return d.SelectOneWithSql(cond, vals)
}

func (d *Dao) SelectMulti(where map[string]interface{}, useSlave ...bool) []Modeller {
	cond, vals, err := builder.BuildSelect(d.GetTableName(), where, nil)
	d.CheckError(err)
	return d.SelectMultiWithSql(cond, vals)
}

func (d *Dao) SelectOneWithSql(query string, params []interface{}, useSlave ...bool) Modeller {
	var (
		row *sql.Rows
		err error
	)
	if len(useSlave) > 0 && !useSlave[0] {
		row, err = d.GetDaoSession().Query(query, params...)
	} else {
		row, err = db.GetSlaveInstance().QueryContext(d.GetDaoSession().ctx, query, params...)
	}
	d.CheckError(err)
	return d.createOneFromRows(row)
}

func (d *Dao) SelectMultiWithSql(query string, params []interface{}, useSlave ...bool) []Modeller {
	var (
		row *sql.Rows
		err error
	)
	if len(useSlave) > 0 && !useSlave[0] {
		row, err = d.GetDaoSession().Query(query, params...)
	} else {
		row, err = db.GetSlaveInstance().QueryContext(d.GetDaoSession().ctx, query, params...)
	}
	d.CheckError(err)
	return d.createMultiFromRows(row)
}

func (d *Dao) ExecWithSql(query string, params []interface{}) sql.Result {
	result, err := d.GetDaoSession().Exec(query, params...)
	d.CheckError(err)
	return result
}

func (d *Dao) ResolveDataFromRows(rows *sql.Rows) []map[string]interface{} {
	columns, err := rows.Columns()
	d.CheckError(err)
	length := len(columns)
	values := make([]interface{}, length, length)
	for i := 0; i < length; i++ {
		values[i] = new(interface{})
	}
	var data = make([]map[string]interface{}, 0)
	for rows.Next() {
		err = rows.Scan(values...)
		d.CheckError(err)
		mp := make(map[string]interface{})
		for idx, name := range columns {
			mp[name] = *(values[idx].(*interface{}))
		}
		data = append(data, mp)
	}
	return data
}
