package sorm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/xkisas/sorm/builder"
	"github.com/xkisas/sorm/db"
	"github.com/xkisas/sorm/internal"
)

const defaultTagName = "db"

type DaoIfe interface {
	initDao(dao DaoIfe, tableName string, indexFields, fields []string, session *Session, modelType reflect.Type, notFoundError error)
	buildWhere(indexes ...interface{}) (map[string]interface{}, error)
	update(model ModelIfe, data map[string]interface{}) (int64, error)
	remove(model ModelIfe) error
	Session() *Session
	GetTableName() string
	Insert(data map[string]interface{}, indexValues ...interface{}) (model ModelIfe, err error)
	Select(forUpdate bool, indexValues ...interface{}) (ModelIfe, error)
	SelectById(id interface{}, opts ...Option) (ModelIfe, error)
	SelectOne(where map[string]interface{}, opts ...Option) (ModelIfe, error)
	SelectOneWithSql(query string, params []interface{}, opts ...Option) (ModelIfe, error)
	SelectMulti(where map[string]interface{}, opts ...Option) ([]ModelIfe, error)
	SelectMultiWithSql(query string, params []interface{}, opts ...Option) ([]ModelIfe, error)
	GetCount(column string, where map[string]interface{}, opts ...Option) (int, error)
	GetSum(column string, where map[string]interface{}, opts ...Option) (int, error)
	ExecWithSql(query string, params []interface{}) (sql.Result, error)
	QueryWithSql(query string, params []interface{}, opts ...Option) (*sql.Rows, error)
	ResolveModelFromRows(rows *sql.Rows) ([]ModelIfe, error)
}

type Dao struct {
	customDao     DaoIfe
	tableName     string       // dao绑定的表
	indexFields   []string     // 主键字段
	fields        []string     // 表字段
	notFoundError error        // 记录未找到时报错
	session       *Session     // 绑定session
	modelType     reflect.Type // 通过反射可用于构造model对象

	locker sync.Mutex
}

var (
	ModelRuntimeError  = errors.New("model runtime error")
	ModelNotFoundError = errors.New("model not found error")
)

func (d *Dao) initDao(dao DaoIfe, tableName string, indexFields, fields []string, session *Session, modelType reflect.Type, notFoundError error) {
	d.customDao = dao
	d.tableName = tableName
	d.indexFields = indexFields
	d.fields = fields
	d.session = session
	d.modelType = modelType
	d.notFoundError = notFoundError
}

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
	indexValues := make([]interface{}, 0, len(d.indexFields))
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
func (d *Dao) CreateObj(data map[string]interface{}, loaded bool, indexValues ...interface{}) (ModelIfe, error) {
	var (
		model ModelIfe
		ok    bool
		err   error
	)
	var indexValuesCopy []interface{}
	if len(indexValues) == 0 {
		if indexValuesCopy, ok = d.getIndexValuesFromData(data); !ok {
			return nil, NewError(ModelRuntimeError, "index values not found")
		}
	} else {
		indexValuesCopy = make([]interface{}, len(indexValues))
		copy(indexValuesCopy, indexValues)
	}

	d.locker.Lock()
	defer d.locker.Unlock()
	if model, err = d.QueryCache(indexValuesCopy...); err == nil && model != nil && !loaded {
		return model, nil
	}
	if model == nil {
		vc := reflect.New(d.modelType)
		model = vc.Interface().(ModelIfe)
	}

	if err = internal.ScanStruct(data, model, defaultTagName, true); err != nil {
		return nil, err
	}
	model.initBase(d.customDao, indexValuesCopy, loaded)
	d.SaveCache(model)
	return model, nil
}

func (d *Dao) update(model ModelIfe, data map[string]interface{}) (int64, error) {
	where, err := d.buildWhere(model.IndexValues()...)
	if err != nil {
		return 0, err
	}
	query, params, err := builder.Update().Table(d.GetTableName()).Set(data).Where(where).Build()
	if err != nil {
		return 0, err
	}
	result, err := d.ExecWithSql(query, params)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if affected == 1 {
		if err = internal.ScanStruct(data, model, defaultTagName, true); err != nil {
			return affected, err
		}

		d.SaveCache(model)
	}
	return affected, err
}

func (d *Dao) remove(model ModelIfe) error {
	indexValues := model.IndexValues()
	where, err := d.buildWhere(indexValues...)
	if err != nil {
		return err
	}
	query, params, err := builder.Delete().Table(d.GetTableName()).Where(where).Build()
	if err != nil {
		return err
	}
	result, err := d.ExecWithSql(query, params)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return err
	} else if affected == 0 {
		return d.notFoundError
	}
	d.RemoveCache(indexValues...)
	return nil
}

func (d *Dao) Session() *Session {
	return d.session
}

func (d *Dao) GetTableName() string {
	return d.tableName
}

func (d *Dao) Insert(data map[string]interface{}, indexValues ...interface{}) (model ModelIfe, err error) {
	var pk = make([]interface{}, 0)
	if len(indexValues) > 0 {
		if len(indexValues) != len(d.indexFields) {
			return nil, NewError(ModelRuntimeError, "indexValues num error")
		}
		for i, index := range indexValues {
			data[d.indexFields[i]] = index
			pk = append(pk, index)
		}
	}
	query, params, err := builder.Insert().Table(d.GetTableName()).Values(data).Build()
	if err != nil {
		return nil, err
	}
	err = d.Session().runInTransaction(func() error {
		result, err := d.ExecWithSql(query, params)
		if err != nil {
			return err
		}
		if affected, err := result.RowsAffected(); err != nil {
			return err
		} else if affected != 1 {
			return NewError(ModelRuntimeError, "dao.baseInsert error")
		}
		if len(pk) == 0 && len(d.indexFields) == 1 {
			if id, err := result.LastInsertId(); err == nil {
				data[d.indexFields[0]] = id
				pk = append(pk, id)
			}
		}
		model, err = d.CreateObj(data, false, pk...)
		return err
	})
	return
}

func (d *Dao) Select(forUpdate bool, indexValues ...interface{}) (ModelIfe, error) {
	if forUpdate {
		where, err := d.buildWhere(indexValues...)
		if err != nil {
			return nil, err
		}
		query, params, err := builder.Select().Table(d.GetTableName()).Columns(d.fields...).Where(where).Tail("FOR UPDATE").Build()
		if err != nil {
			return nil, err
		}

		sess := d.Session()
		sess.txMutex.RLock()
		defer sess.txMutex.RUnlock()

		if sess.tx == nil {
			return nil, NewError(ModelRuntimeError, "Attempt to load for update out of transaction")
		}
		row, err := sess.Query(query, params...)
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
	if obj, err := d.QueryCache(indexValues...); err == nil {
		return obj, nil
	}
	where, err := d.buildWhere(indexValues...)
	if err != nil {
		return nil, err
	}
	return d.CreateObj(where, false)
}

func (d *Dao) SelectById(id interface{}, opts ...Option) (ModelIfe, error) {
	option := fetchOption(opts...)
	model, err := d.Select(option.forUpdate, id)
	if err != nil {
		return nil, err
	}
	if !option.forUpdate && (option.forceLoad || option.load) {
		return model.Load(opts...)
	}
	return model, nil
}

func (d *Dao) SelectOne(where map[string]interface{}, opts ...Option) (ModelIfe, error) {
	query, params, err := builder.Select().Table(d.GetTableName()).Columns(d.fields...).Where(where).Build()
	if err != nil {
		return nil, err
	}
	return d.SelectOneWithSql(query, params, opts...)
}

func (d *Dao) SelectOneWithSql(query string, params []interface{}, opts ...Option) (ModelIfe, error) {
	row, err := d.QueryWithSql(query, params, opts...)
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

func (d *Dao) SelectMulti(where map[string]interface{}, opts ...Option) ([]ModelIfe, error) {
	query, params, err := builder.Select().Table(d.GetTableName()).Columns(d.fields...).Where(where).Build()
	if err != nil {
		return nil, err
	}
	return d.SelectMultiWithSql(query, params, opts...)
}

func (d *Dao) SelectMultiWithSql(query string, params []interface{}, opts ...Option) ([]ModelIfe, error) {
	rows, err := d.QueryWithSql(query, params, opts...)
	if err != nil {
		return nil, err
	}
	return d.ResolveModelFromRows(rows)
}

func (d *Dao) SelectChan(query string, params []interface{}, opts ...Option) (<-chan ModelIfe, <-chan error) {
	var (
		modelCh = make(chan ModelIfe, 0)
		errCh   = make(chan error, 0)
	)
	go func(d *Dao, modelCh chan<- ModelIfe, errCh chan<- error) {
		defer close(modelCh)
		defer close(errCh)
		rows, err := d.QueryWithSql(query, params, opts...)
		if err != nil {
			errCh <- err
			return
		}
		defer rows.Close()
		columns, err := rows.Columns()
		if err != nil {
			errCh <- err
			return
		}
		length := len(columns)
		values := make([]interface{}, length, length)
		for i := 0; i < length; i++ {
			values[i] = new(interface{})
		}
		for rows.Next() {
			err = rows.Scan(values...)
			if err != nil {
				errCh <- err
				return
			}
			mp := make(map[string]interface{})
			for idx, name := range columns {
				mp[name] = *(values[idx].(*interface{}))
			}
			if m, err := d.CreateObj(mp, true); err == nil {
				modelCh <- m
			} else {
				errCh <- err
				return
			}
		}
	}(d, modelCh, errCh)
	return modelCh, errCh
}

func (d *Dao) GetCount(column string, where map[string]interface{}, opts ...Option) (int, error) {
	query, params, err := builder.Select().Table(d.GetTableName()).FuncColumns(map[string]string{
		fmt.Sprintf("COUNT(%s)", builder.QuoteIdentifier(column)): "c",
	}).Where(where).Build()
	if err != nil {
		return 0, err
	}
	rows, err := d.QueryWithSql(query, params, opts...)
	if err != nil {
		return 0, err
	}
	var result struct {
		C int `count:"c"`
	}
	err = ResolveFromRows(rows, &result, "count")
	if err != nil {
		return 0, err
	}
	return result.C, nil
}

func (d *Dao) GetSum(column string, where map[string]interface{}, opts ...Option) (int, error) {
	query, params, err := builder.Select().Table(d.GetTableName()).FuncColumns(map[string]string{
		fmt.Sprintf("SUM(%s)", builder.QuoteIdentifier(column)): "s",
	}).Where(where).Build()
	if err != nil {
		return 0, err
	}
	rows, err := d.QueryWithSql(query, params, opts...)
	if err != nil {
		return 0, err
	}
	var result struct {
		S int `sum:"s"`
	}
	err = ResolveFromRows(rows, &result, "sum")
	if err != nil {
		return 0, err
	}
	return result.S, nil
}

func (d *Dao) ExecWithSql(query string, params []interface{}) (sql.Result, error) {
	return d.Session().Exec(query, params...)
}

func (d *Dao) QueryWithSql(query string, params []interface{}, opts ...Option) (*sql.Rows, error) {
	option := fetchOption(opts...)
	replicaInstance := db.GetReplicaInstance()
	if option.forceMaster || replicaInstance == nil {
		return d.Session().Query(query, params...)
	} else {
		return replicaInstance.QueryContext(d.Session().ctx, query, params...)
	}
}

func (d *Dao) BuildSelect() *builder.Selector {
	return builder.Select().Table(d.GetTableName())
}

func (d *Dao) BuildUpdate() *builder.Updater {
	return builder.Update().Table(d.GetTableName())
}

func (d *Dao) BuildInsert() *builder.Inserter {
	return builder.Insert().Table(d.GetTableName())
}

func (d *Dao) BuildDelete() *builder.Deleter {
	return builder.Delete().Table(d.GetTableName())
}

func (d *Dao) ResolveChan(modelCh <-chan ModelIfe, errCh <-chan error, f func(model ModelIfe) error) error {
	breakLoop := false
	for !breakLoop {
		select {
		case err := <-errCh:
			if err != nil {
				return err
			}
			breakLoop = true
		case model := <-modelCh:
			if model == nil {
				breakLoop = true
			} else if err := f(model); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *Dao) ResolveModelFromRows(rows *sql.Rows) ([]ModelIfe, error) {
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
	var data = make([]ModelIfe, 0)
	for rows.Next() {
		err = rows.Scan(values...)
		if err != nil {
			return nil, err
		}
		mp := make(map[string]interface{})
		for idx, name := range columns {
			mp[name] = *(values[idx].(*interface{}))
		}
		if m, err := d.CreateObj(mp, true); err == nil {
			data = append(data, m)
		} else {
			return nil, err
		}
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
		if len(data) > 0 {
			err = internal.ScanStructSlice(data, target, tagName, false)
		}
	default:
		if len(data) > 0 {
			err = internal.ScanStruct(data[0], target, tagName, false)
		} else {
			err = NewError(ModelRuntimeError, "empty data")
		}
	}
	return err
}
