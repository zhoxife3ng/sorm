package sorm

import (
	"database/sql"
	"errors"
	"github.com/x554462/sorm/builder"
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
	notFoundError error
	// 绑定session
	session *Session
	// 通过反射可用于构造model对象
	modelType reflect.Type
}

var (
	ModelRuntimeError  = errors.New("model runtime error")
	ModelNotFoundError = errors.New("model not found error")
)

func (d *Dao) buildWhere(indexes ...interface{}) (map[string]interface{}, error) {
	if len(d.indexFields) != len(indexes) {
		return nil, NewError(ModelRuntimeError, "dao.buildWhere index number error")
	}
	where := make(map[string]interface{})
	for i, v := range d.indexFields {
		where[v] = indexes[i]
	}
	return where, nil
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
func (d *Dao) createOne(data map[string]interface{}, indexValues []interface{}, loaded bool) (Modeller, error) {
	var (
		model Modeller
		ok    bool
		err   error
	)
	if indexValues, ok = d.getIndexValuesFromData(data); ok {
		if model, err = d.queryCache(indexValues...); err == nil && model != nil && !loaded {
			return model, nil
		}
	}
	if model == nil {
		vc := reflect.New(d.modelType)
		model = vc.Interface().(Modeller)
	}
	if err = internal.ScanStruct(data, model, defaultTagName); err != nil {
		return nil, err
	}
	model.initBase(d, indexValues, loaded)
	d.saveCache(model)
	return model, nil
}

func (d *Dao) update(model Modeller, data map[string]interface{}) (int64, error) {
	where, err := d.buildWhere(model.IndexValues()...)
	if err != nil {
		return 0, err
	}
	cond, params, err := builder.Update().Table(d.GetTableName()).Set(data).Where(where).Build()
	if err != nil {
		return 0, err
	}
	result, err := d.ExecWithSql(cond, params)
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
	where, err := d.buildWhere(indexValues...)
	if err != nil {
		return err
	}
	cond, params, err := builder.Delete().Table(d.GetTableName()).Where(where).Build()
	if err != nil {
		return err
	}
	result, err := d.ExecWithSql(cond, params)
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

func (d *Dao) Select(forUpdate bool, indexValues ...interface{}) (Modeller, error) {
	if forUpdate {
		where, err := d.buildWhere(indexValues...)
		if err != nil {
			return nil, err
		}
		cond, params, err := builder.Select().Table(d.GetTableName()).Columns(d.fields...).Where(where).Tail("FOR UPDATE").Build()
		if err != nil {
			return nil, err
		}
		if d.Session().tx == nil {
			return nil, NewError(ModelRuntimeError, "Attempt to load for update out of transaction")
		}
		row, err := d.Session().Query(cond, params...)
		if err != nil {
			return nil, err
		}
		ms, err := d.ResolveModelFromRows(row)
		if err != nil {
			return nil, err
		} else if len(ms) < 1 {
			return nil, d.notFoundError
		}
		return ms[0], nil
	}
	if obj, err := d.queryCache(indexValues...); err == nil {
		return obj, nil
	}
	where, err := d.buildWhere(indexValues...)
	if err != nil {
		return nil, err
	}
	return d.createOne(where, indexValues, false)
}

func (d *Dao) SelectById(id interface{}, opts ...Option) (Modeller, error) {
	option := newOption()
	for _, o := range opts {
		o(&option)
	}
	model, err := d.Select(option.forUpdate, id)
	if err != nil {
		return nil, err
	}
	if !option.forUpdate && (option.forceLoad || option.load) {
		return model.Load(opts...)
	}
	return model, nil
}

func (d *Dao) Insert(data map[string]interface{}, indexValues ...interface{}) (Modeller, error) {
	cond, params, err := builder.Insert().Table(d.GetTableName()).Values(data).Build()
	if err != nil {
		return nil, err
	}
	result, err := d.ExecWithSql(cond, params)
	if err != nil {
		return nil, err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return nil, err
	} else if affected != 1 {
		return nil, NewError(ModelRuntimeError, "dao.baseInsert error")
	}
	if len(indexValues) > 0 {
		for i, index := range indexValues {
			data[d.indexFields[i]] = index
		}
	} else if len(d.indexFields) == 1 {
		if id, err := result.LastInsertId(); err == nil {
			data[d.indexFields[0]] = id
			indexValues = append(indexValues, id)
		}
	}
	return d.createOne(data, indexValues, false)
}

func (d *Dao) SelectOne(where map[string]interface{}, opts ...Option) (Modeller, error) {
	cond, params, err := builder.Select().Table(d.GetTableName()).Columns(d.fields...).Where(where).Build()
	if err != nil {
		return nil, err
	}
	return d.SelectOneWithSql(cond, params, opts...)
}

func (d *Dao) SelectOneWithSql(query string, params []interface{}, opts ...Option) (Modeller, error) {
	var (
		row    *sql.Rows
		err    error
		option = newOption()
	)
	for _, o := range opts {
		o(&option)
	}
	if option.forceMaster {
		row, err = d.Session().Query(query, params...)
	} else {
		row, err = db.GetReplicaInstance().QueryContext(d.Session().ctx, query, params...)
	}
	if err != nil {
		return nil, err
	}
	ms, err := d.ResolveModelFromRows(row)
	if err != nil {
		return nil, err
	} else if len(ms) < 1 {
		return nil, d.notFoundError
	}
	return ms[0], nil
}

func (d *Dao) SelectMulti(where map[string]interface{}, opts ...Option) ([]Modeller, error) {
	cond, params, err := builder.Select().Table(d.GetTableName()).Columns(d.fields...).Where(where).Build()
	if err != nil {
		return nil, err
	}
	return d.SelectMultiWithSql(cond, params, opts...)
}

func (d *Dao) SelectMultiWithSql(query string, params []interface{}, opts ...Option) ([]Modeller, error) {
	var (
		rows   *sql.Rows
		err    error
		option = newOption()
	)
	for _, o := range opts {
		o(&option)
	}
	if option.forceMaster {
		rows, err = d.Session().Query(query, params...)
	} else {
		rows, err = db.GetReplicaInstance().QueryContext(d.Session().ctx, query, params...)
	}
	if err != nil {
		return nil, err
	}
	return d.ResolveModelFromRows(rows)
}

func (d *Dao) ExecWithSql(query string, params []interface{}) (sql.Result, error) {
	return d.Session().Exec(query, params...)
}

func (d *Dao) ResolveModelFromRows(rows *sql.Rows) ([]Modeller, error) {
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	length := len(columns)
	values := make([]interface{}, length, length)
	for i := 0; i < length; i++ {
		values[i] = new(interface{})
	}
	var (
		data        = make([]Modeller, 0)
		indexValues = make([]interface{}, 0, len(d.indexFields))
	)
	for rows.Next() {
		err = rows.Scan(values...)
		if err != nil {
			return nil, err
		}
		mp := make(map[string]interface{})
		for idx, name := range columns {
			mp[name] = *(values[idx].(*interface{}))
		}
		for _, indexField := range d.indexFields {
			indexValues = append(indexValues, mp[indexField])
		}
		if m, err := d.createOne(mp, indexValues, true); err == nil {
			data = append(data, m)
		} else {
			return nil, err
		}
		indexValues = indexValues[0:0]
	}
	return data, nil
}

func ResolveDataFromRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	length := len(columns)
	values := make([]interface{}, length, length)
	for i := 0; i < length; i++ {
		values[i] = new(interface{})
	}
	var data = make([]map[string]interface{}, 0)
	for rows.Next() {
		err = rows.Scan(values...)
		if err != nil {
			return nil, err
		}
		mp := make(map[string]interface{})
		for idx, name := range columns {
			mp[name] = *(values[idx].(*interface{}))
		}
		data = append(data, mp)
	}
	return data, nil
}

func ResolveFromRows(rows *sql.Rows, target interface{}, tagName string) error {
	data, err := ResolveDataFromRows(rows)
	if err != nil {
		return err
	}
	switch reflect.TypeOf(target).Elem().Kind() {
	case reflect.Slice:
		if len(data) == 0 {
			return nil
		}
		err = internal.ScanStructSlice(data, target, tagName)
	default:
		if len(data) == 0 {
			return ModelNotFoundError
		}
		err = internal.ScanStruct(data[0], target, tagName)
	}
	return err
}
