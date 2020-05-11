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
	lock          sync.RWMutex
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

func (s *Session) GetDao(model Modeller) *Dao {
	t := reflect.Indirect(reflect.ValueOf(model)).Type()
	name := t.Name()

	s.lock.RLock()
	if value, ok := s.daoMap[name]; ok {
		s.lock.RUnlock()
		return value
	}
	s.lock.RUnlock()

	s.lock.Lock()
	defer s.lock.Unlock()
	if value, ok := s.daoMap[name]; ok {
		return value
	}
	tableName, indexFields := parseTableInfo(t)
	if len(indexFields) != len(model.IndexValues()) {
		exception.ThrowMsg("session.GetDao: orm model indexFields error", ModelRuntimeError)
	}
	dao := &Dao{
		tableName:     tableName,
		indexFields:   indexFields,
		notFoundError: model.GetNotFoundError(),
		session:       s,
		modelType:     t,
	}
	s.daoMap[name] = dao
	return dao
}

func (s *Session) BeginTransaction() {
	if s.tx == nil {
		var err error
		if s.tx, err = db.GetInstance().Begin(); err != nil {
			log.Printf(fmt.Sprintf("session.BeginTransaction: %s\n", err.Error()))
		}
	} else {
		log.Printf("session.BeginTransaction: can not begin tx again")
	}
}

func (s *Session) RollbackTransaction() {
	if s.tx != nil {
		if err := s.tx.Rollback(); err != nil {
			log.Printf(fmt.Sprintf("session.RollbackTransaction: %s", err.Error()))
		}
		s.tx = nil
	}
}

func (s *Session) SubmitTransaction() {
	if s.tx != nil {
		if err := s.tx.Commit(); err != nil {
			log.Printf(fmt.Sprintf("session.SubmitTransaction: %s", err.Error()))
		}
		s.tx = nil
	}
}

func (s *Session) InTransaction() bool {
	return s.tx != nil
}

func (s *Session) Close() {
	if s.tx != nil {
		s.RollbackTransaction()
		s.tx = nil
	}
	s.daoMap = make(map[string]*Dao)
	s.daoModelCache.Clear()
	sessionPool.Put(s)
}

func (s *Session) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if s.tx != nil {
		return s.tx.QueryContext(s.ctx, query, args...)
	}
	return db.GetInstance().QueryContext(s.ctx, query, args...)
}

func (s *Session) Exec(query string, args ...interface{}) (sql.Result, error) {
	if s.tx != nil {
		return s.tx.ExecContext(s.ctx, query, args...)
	}
	return db.GetInstance().ExecContext(s.ctx, query, args...)
}

func (s *Session) ClearAllCache() {
	s.daoModelCache.Clear()
}
