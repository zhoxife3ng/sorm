package sorm

import (
	"database/sql"
	"github.com/didi/gendry/builder"
	"github.com/x554462/go-exception"
	"github.com/x554462/sorm/db"
	"github.com/x554462/sorm/internal"
	"reflect"
)

const defaultTagName = "db"

type Dao struct {
	// dao绑定的表
	tableName string
	// 主键字段
	indexFields []string
	// 表字段
	fields []string
	// 记录未找到时报错
	notFoundError exception.ErrorWrapper
	// 绑定session
	session *Session
	// 通过反射可用于构造model对象
	modelType reflect.Type
}

var (
	ModelRuntimeError  = exception.New("model runtime error", exception.RootError)
	ModelNotFoundError = exception.New("model not found error", exception.RootError)
)

func (d *Dao) checkError(err error) {
	if err != nil {
		exception.ThrowMsgWithCallerDepth(err.Error(), ModelRuntimeError, 3)
	}
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

// 创建model对象
func (d *Dao) createOne(data map[string]interface{}, loaded bool) Modeller {
	if indexValues, ok := d.getIndexValuesFromData(data); ok {
		if model := d.queryCache(indexValues...); model != nil {
			return model
		}
	}
	vc := reflect.New(d.modelType)
	model, ok := vc.Interface().(Modeller)
	if !ok {
		exception.ThrowMsg("dao.newModel error", ModelRuntimeError)
	}
	d.checkError(internal.ScanStruct(data, model, defaultTagName))
	model.initBase(d, model, loaded)
	d.saveCache(model)
	return model
}

func (d *Dao) update(model Modeller, data map[string]interface{}) (int64, error) {
	cond, vals, err := builder.BuildUpdate(d.selectTableName(), d.buildWhere(model.IndexValues()...), data)
	if err != nil {
		return 0, err
	}
	result, err := d.ExecWithSql(cond, vals)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if affected == 1 {
		internal.ScanStruct(data, model, defaultTagName)
		d.saveCache(model)
	}
	return affected, err
}

func (d *Dao) remove(model Modeller) error {
	indexValues := model.IndexValues()
	cond, vals, err := builder.BuildDelete(d.selectTableName(), d.buildWhere(indexValues...))
	if err != nil {
		return err
	}
	result, err := d.ExecWithSql(cond, vals)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return err
	} else if affected == 0 {
		return d.notFoundError
	}
	d.removeCache(indexValues...)
	return nil
}

func (d *Dao) Session() *Session {
	return d.session
}

func (d *Dao) GetTableName() string {
	return d.tableName
}

func (d *Dao) selectTableName() string {
	return "`" + d.tableName + "`"
}

func (d *Dao) Select(forUpdate bool, indexes ...interface{}) (Modeller, error) {
	if forUpdate {
		cond, vals, err := builder.BuildSelect(d.selectTableName(), d.buildWhere(indexes...), d.fields)
		d.checkError(err)
		if d.Session().tx == nil {
			return nil, exception.New("Attempt to load for update out of transaction", ModelRuntimeError)
		}
		cond = cond + " FOR UPDATE"
		row, err := d.Session().Query(cond, vals...)
		if err != nil {
			return nil, err
		}
		ms := d.ResolveModelFromRows(row)
		if len(ms) < 1 {
			return nil, d.notFoundError
		}
		return ms[0], nil
	}
	obj := d.queryCache(indexes...)
	if obj != nil {
		return obj, nil
	}
	return d.createOne(d.buildWhere(indexes...), false), nil
}

func (d *Dao) SelectById(id interface{}, opts ...option) (Modeller, error) {
	options := newOptions()
	for _, o := range opts {
		o(&options)
	}
	model, err := d.Select(options.forUpdate, id)
	if err != nil {
		return nil, err
	}
	if !options.forUpdate && (options.forceLoad || options.load) {
		return model.Load(opts...)
	}
	return model, nil
}

func (d *Dao) Insert(data map[string]interface{}, indexes ...interface{}) (Modeller, error) {
	cond, vals, err := builder.BuildInsert(d.selectTableName(), []map[string]interface{}{data})
	d.checkError(err)
	result, err := d.ExecWithSql(cond, vals)
	if err != nil {
		return nil, err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return nil, err
	} else if affected != 1 {
		return nil, exception.New("dao.Insert error", ModelRuntimeError)
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
	return d.createOne(data, false), nil
}

func (d *Dao) SelectOne(where map[string]interface{}, opts ...option) (Modeller, error) {
	cond, vals, err := builder.BuildSelect(d.selectTableName(), where, d.fields)
	d.checkError(err)
	return d.SelectOneWithSql(cond, vals, opts...)
}

func (d *Dao) SelectOneWithSql(query string, params []interface{}, opts ...option) (Modeller, error) {
	var (
		row     *sql.Rows
		err     error
		options = newOptions()
	)
	for _, o := range opts {
		o(&options)
	}
	if options.forceMaster {
		row, err = d.Session().Query(query, params...)
	} else {
		row, err = db.GetSlaveInstance().QueryContext(d.Session().ctx, query, params...)
	}
	if err != nil {
		return nil, err
	}
	ms := d.ResolveModelFromRows(row)
	if len(ms) < 1 {
		return nil, d.notFoundError
	}
	return ms[0], nil
}

func (d *Dao) SelectMulti(where map[string]interface{}, opts ...option) []Modeller {
	cond, vals, err := builder.BuildSelect(d.selectTableName(), where, d.fields)
	d.checkError(err)
	return d.SelectMultiWithSql(cond, vals, opts...)
}

func (d *Dao) SelectMultiWithSql(query string, params []interface{}, opts ...option) []Modeller {
	var (
		rows    *sql.Rows
		err     error
		options = newOptions()
	)
	for _, o := range opts {
		o(&options)
	}
	if options.forceMaster {
		rows, err = d.Session().Query(query, params...)
	} else {
		rows, err = db.GetSlaveInstance().QueryContext(d.Session().ctx, query, params...)
	}
	d.checkError(err)
	return d.ResolveModelFromRows(rows)
}

func (d *Dao) ExecWithSql(query string, params []interface{}) (sql.Result, error) {
	return d.Session().Exec(query, params...)
}

func (d *Dao) ResolveModelFromRows(rows *sql.Rows) []Modeller {
	defer rows.Close()
	columns, err := rows.Columns()
	d.checkError(err)
	length := len(columns)
	values := make([]interface{}, length, length)
	for i := 0; i < length; i++ {
		values[i] = new(interface{})
	}
	var data = make([]Modeller, 0)
	for rows.Next() {
		err = rows.Scan(values...)
		d.checkError(err)
		mp := make(map[string]interface{})
		for idx, name := range columns {
			mp[name] = *(values[idx].(*interface{}))
		}
		data = append(data, d.createOne(mp, true))
	}
	return data
}

func (d *Dao) ResolveDataFromRows(rows *sql.Rows) []map[string]interface{} {
	defer rows.Close()
	columns, err := rows.Columns()
	d.checkError(err)
	length := len(columns)
	values := make([]interface{}, length, length)
	for i := 0; i < length; i++ {
		values[i] = new(interface{})
	}
	var data = make([]map[string]interface{}, 0)
	for rows.Next() {
		err = rows.Scan(values...)
		d.checkError(err)
		mp := make(map[string]interface{})
		for idx, name := range columns {
			mp[name] = *(values[idx].(*interface{}))
		}
		data = append(data, mp)
	}
	return data
}
