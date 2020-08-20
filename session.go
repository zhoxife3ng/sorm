package sorm

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/x554462/sorm/db"
	"log"
	"reflect"
	"sync"
)

const daoModelLruCacheSize = 50

var sessionPool = sync.Pool{
	New: func() interface{} {
		return &Session{daoModelCache: newDaoLru(daoModelLruCacheSize), daoMap: make(map[string]DaoIfe)}
	},
}

type Session struct {
	daoLocker     sync.RWMutex
	txLocker      sync.Mutex
	tx            *sql.Tx
	daoMap        map[string]DaoIfe
	daoModelCache *modelLruCache
	ctx           context.Context
}

func NewSession(ctx context.Context) *Session {
	sess := sessionPool.Get().(*Session)
	sess.ctx = ctx
	return sess
}

func (s *Session) GetDao(model ModelIfe) DaoIfe {
	t := reflect.Indirect(reflect.ValueOf(model)).Type()
	name := t.Name()

	s.daoLocker.RLock()
	if value, ok := s.daoMap[name]; ok {
		s.daoLocker.RUnlock()
		return value
	}
	s.daoLocker.RUnlock()

	s.daoLocker.Lock()
	defer s.daoLocker.Unlock()
	if value, ok := s.daoMap[name]; ok {
		return value
	}
	tableName, indexFields, fields := parseTableInfo(t)
	dao := model.CustomDao()
	dao.initDao(dao, tableName, indexFields, fields, s, t, model.GetNotFoundError())
	s.daoMap[name] = dao
	return dao
}

func (s *Session) BeginTransaction() {
	s.txLocker.Lock()
	defer s.txLocker.Unlock()
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
	s.txLocker.Lock()
	defer s.txLocker.Unlock()
	if s.tx != nil {
		if err := s.tx.Rollback(); err != nil {
			log.Printf(fmt.Sprintf("session.RollbackTransaction: %s", err.Error()))
		}
		s.tx = nil
	}
}

func (s *Session) SubmitTransaction() {
	s.txLocker.Lock()
	defer s.txLocker.Unlock()
	if s.tx != nil {
		if err := s.tx.Commit(); err != nil {
			log.Printf(fmt.Sprintf("session.SubmitTransaction: %s", err.Error()))
		}
		s.tx = nil
	}
}

func (s *Session) InTransaction() bool {
	s.txLocker.Lock()
	defer s.txLocker.Unlock()
	return s.tx != nil
}

func (s *Session) Close() {
	s.txLocker.Lock()
	defer s.txLocker.Unlock()
	if s.tx != nil {
		s.RollbackTransaction()
		s.tx = nil
	}
	s.daoMap = make(map[string]DaoIfe)
	s.daoModelCache.Clear()
	s.ctx = nil
	sessionPool.Put(s)
}

func (s *Session) QueryReplica(query string, args ...interface{}) (*sql.Rows, error) {
	return db.GetReplicaInstance().QueryContext(s.ctx, query, args...)
}

func (s *Session) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if s.InTransaction() {
		return s.tx.QueryContext(s.ctx, query, args...)
	}
	return db.GetInstance().QueryContext(s.ctx, query, args...)
}

func (s *Session) Exec(query string, args ...interface{}) (sql.Result, error) {
	if s.InTransaction() {
		return s.tx.ExecContext(s.ctx, query, args...)
	}
	return db.GetInstance().ExecContext(s.ctx, query, args...)
}

func (s *Session) RunInTransaction(f func() error) (err error) {
	if s.InTransaction() {
		err = f()
	} else {
		s.BeginTransaction()
		err = f()
		if err != nil {
			s.RollbackTransaction()
		} else {
			s.SubmitTransaction()
		}
	}
	return
}

func (s *Session) ClearAllCache() {
	s.daoModelCache.Clear()
}
