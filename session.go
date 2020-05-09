package sorm

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/x554462/go-exception"
	"github.com/x554462/sorm/db"
	"log"
	"reflect"
	"sync"
)

const daoModelLruCacheSize = 50

var sessionPool = sync.Pool{
	New: func() interface{} {
		return &Session{daoModelCache: newDaoLru(daoModelLruCacheSize)}
	},
}

type Session struct {
	mu            sync.Mutex
	tx            *sql.Tx
	daoMap        map[string]*Dao
	daoModelCache *modelLruCache
	ctx           context.Context
}

func NewSession(ctx context.Context) *Session {
	sess := sessionPool.Get().(*Session)
	sess.daoMap = make(map[string]*Dao)
	sess.ctx = ctx
	return sess
}

func (ds *Session) Get(model Modeller) *Dao {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	t := reflect.Indirect(reflect.ValueOf(model)).Type()
	name := t.Name()
	if value, ok := ds.daoMap[name]; ok {
		return value
	}
	tableName, indexFields := parseTableInfo(t)
	if len(indexFields) != len(model.IndexValues()) {
		exception.ThrowMsg("dao.initDao: model indexFields error", ModelRuntimeError)
	}
	dao := &Dao{
		tableName:     tableName,
		indexFields:   indexFields,
		notFoundError: model.GetNotFoundError(),
		daoSession:    ds,
		modelType:     t,
	}
	ds.daoMap[name] = dao
	return dao
}

func (ds *Session) BeginTransaction() {
	if ds.tx == nil {
		var err error
		if ds.tx, err = db.GetInstance().Begin(); err != nil {
			log.Printf(fmt.Sprintf("session.BeginTransaction: %s\n", err.Error()))
		}
	} else {
		log.Printf("session.BeginTransaction: can not begin tx again")
	}
}

func (ds *Session) RollbackTransaction() {
	if ds.tx != nil {
		if err := ds.tx.Rollback(); err != nil {
			log.Printf(fmt.Sprintf("session.RollbackTransaction: %s", err.Error()))
		}
		ds.tx = nil
	}
}

func (ds *Session) SubmitTransaction() {
	if ds.tx != nil {
		if err := ds.tx.Commit(); err != nil {
			log.Printf(fmt.Sprintf("session.SubmitTransaction: %s", err.Error()))
		}
		ds.tx = nil
	}
}

func (ds *Session) InTransaction() bool {
	if ds.tx == nil {
		return false
	}
	return true
}

func (ds *Session) Close() {
	if ds.tx != nil {
		ds.RollbackTransaction()
		ds.tx = nil
	}
	ds.daoMap = make(map[string]*Dao)
	ds.daoModelCache.Clear()
	sessionPool.Put(ds)
}

func (ds *Session) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if ds.tx != nil {
		return ds.tx.QueryContext(ds.ctx, query, args...)
	}
	return db.GetInstance().QueryContext(ds.ctx, query, args...)
}

func (ds *Session) Exec(query string, args ...interface{}) (sql.Result, error) {
	if ds.tx != nil {
		return ds.tx.ExecContext(ds.ctx, query, args...)
	}
	return db.GetInstance().ExecContext(ds.ctx, query, args...)
}

func (ds *Session) ClearAllCache() {
	ds.daoModelCache.Clear()
}
