package sorm

import (
	"context"
	"database/sql"
	"log"
	"reflect"
	"sync"

	"github.com/xkisas/sorm/db"
)

const daoModelLruCacheSize = 200

type Session struct {
	tx            *sql.Tx
	txMutex       sync.RWMutex
	daoMap        map[string]DaoIfe
	daoMapLocker  sync.RWMutex
	daoModelCache *modelLruCache
	ctx           context.Context
}

var sessionPool = sync.Pool{
	New: func() interface{} {
		return &Session{daoModelCache: newDaoLru(daoModelLruCacheSize), daoMap: make(map[string]DaoIfe)}
	},
}

func NewSession(ctx context.Context) *Session {
	sess := sessionPool.Get().(*Session)
	sess.ctx = ctx
	return sess
}

func (s *Session) NewSession() *Session {
	return NewSession(s.ctx)
}

func (s *Session) GetDao(model ModelIfe) DaoIfe {
	t := reflect.Indirect(reflect.ValueOf(model)).Type()
	name := t.Name()

	s.daoMapLocker.RLock()
	if value, ok := s.daoMap[name]; ok {
		s.daoMapLocker.RUnlock()
		return value
	}
	s.daoMapLocker.RUnlock()

	s.daoMapLocker.Lock()
	defer s.daoMapLocker.Unlock()
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
	s.txMutex.Lock()
	defer s.txMutex.Unlock()
	s.txBegin()
}

func (s *Session) RollbackTransaction() {
	s.txMutex.Lock()
	defer s.txMutex.Unlock()
	s.txRollback()
}

func (s *Session) SubmitTransaction() {
	s.txMutex.Lock()
	defer s.txMutex.Unlock()
	s.txCommit()
}

func (s *Session) InTransaction() bool {
	s.txMutex.RLock()
	defer s.txMutex.RUnlock()
	return s.tx != nil
}

func (s *Session) Close() {
	s.RollbackTransaction()
	s.daoMap = make(map[string]DaoIfe)
	s.daoModelCache.Clear()
	s.ctx = nil
	sessionPool.Put(s)
}

func (s *Session) QueryReplica(query string, args ...interface{}) (*sql.Rows, error) {
	replicaInstance := db.GetReplicaInstance()
	if replicaInstance == nil {
		return nil, NewError(ModelRuntimeError, "replica instance is nil")
	}
	return replicaInstance.QueryContext(s.ctx, query, args...)
}

func (s *Session) Query(query string, args ...interface{}) (rows *sql.Rows, err error) {
	s.txMutex.RLock()
	defer s.txMutex.RUnlock()
	if s.tx != nil {
		rows, err = s.tx.QueryContext(s.ctx, query, args...)
	} else {
		rows, err = db.GetInstance().QueryContext(s.ctx, query, args...)
	}
	return
}

func (s *Session) Exec(query string, args ...interface{}) (result sql.Result, err error) {
	s.txMutex.RLock()
	defer s.txMutex.RUnlock()
	if s.tx != nil {
		result, err = s.tx.ExecContext(s.ctx, query, args...)
	} else {
		result, err = db.GetInstance().ExecContext(s.ctx, query, args...)
	}
	return
}

func (s *Session) ClearAllCache() {
	s.daoModelCache.Clear()
}

func (s *Session) runInTransaction(f func() error) (err error) {
	s.txMutex.RLock()
	defer s.txMutex.RUnlock()
	if s.tx != nil {
		err = f()
	} else {
		s.txBegin()
		err = f()
		if err != nil {
			s.txRollback()
		} else {
			s.txCommit()
		}
	}
	return
}

func (s *Session) txBegin() {
	if s.tx != nil {
		log.Printf("session.txBegin: can not begin tx again")
	} else {
		var err error
		if s.tx, err = db.GetInstance().Begin(); err != nil {
			log.Printf("session.txBegin: %s\n", err.Error())
		}
	}
}

func (s *Session) txRollback() {
	if s.tx != nil {
		if err := s.tx.Rollback(); err != nil {
			log.Printf("session.txRollback: %s\n", err.Error())
		}
		s.tx = nil
	}
}

func (s *Session) txCommit() {
	if s.tx != nil {
		if err := s.tx.Commit(); err != nil {
			log.Printf("session.txCommit: %s\n", err.Error())
		}
		s.tx = nil
	}
}
