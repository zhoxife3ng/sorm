package sorm

import (
	"context"
	"database/sql"
	"log"
	"reflect"
	"sync"

	"github.com/xkisas/sorm/db"
)

var daoModelLruCacheCapacity = 200

type Session struct {
	tx            *sql.Tx
	txMutex       sync.RWMutex
	daoMap        map[string]DaoIfe
	daoMapLocker  sync.RWMutex
	daoModelCache *modelLruCache
	ctx           context.Context
	logSql        bool
}

var sessionPool = sync.Pool{
	New: func() interface{} {
		return &Session{daoModelCache: newDaoLru(daoModelLruCacheCapacity)}
	},
}

func SetCacheCapacity(capacity int) {
	daoModelLruCacheCapacity = capacity
}

func NewSession(ctx context.Context) *Session {
	sess := sessionPool.Get().(*Session)
	sess.ctx = ctx
	sess.logSql = false
	sess.daoMap = make(map[string]DaoIfe)
	return sess
}

func (s *Session) NewSession() *Session {
	return NewSession(s.ctx)
}

func (s *Session) SetLogSql(b bool) *Session {
	s.logSql = b
	return s
}

func (s *Session) log(prefix, query string, args []interface{}) {
	if s.logSql {
		log.Println(prefix, query, args)
	}
}

func (s *Session) ResetCacheCapacity(capacity int) {
	s.daoModelCache.resetCapacity(capacity)
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
	dao := model.InitCustomDao()
	if dao == nil {
		if cd, ok := customDaoMap.Load(name); ok {
			dao = reflect.New(cd.(reflect.Type)).Interface().(DaoIfe)
		} else {
			dao = new(Dao)
		}
	}
	tableName, indexFields, fields := parseTableInfo(t)
	dao.initDao(dao, tableName, indexFields, fields, s, t, model.GetNotFoundError())
	s.daoMap[name] = dao
	return dao
}

func (s *Session) BeginTransaction() error {
	s.txMutex.Lock()
	defer s.txMutex.Unlock()
	return s.txBegin()
}

func (s *Session) RollbackTransaction() error {
	s.txMutex.Lock()
	defer s.txMutex.Unlock()
	return s.txRollback()
}

func (s *Session) SubmitTransaction() error {
	s.txMutex.Lock()
	defer s.txMutex.Unlock()
	return s.txCommit()
}

func (s *Session) InTransaction() bool {
	s.txMutex.RLock()
	defer s.txMutex.RUnlock()
	return s.tx != nil
}

func (s *Session) Close() {
	s.daoMapLocker.Lock()
	defer s.daoMapLocker.Unlock()

	s.RollbackTransaction()
	s.daoMap = nil
	s.daoModelCache.Clear()
	sessionPool.Put(s)
}

func (s *Session) QueryReplica(query string, args ...interface{}) (*sql.Rows, error) {
	replicaInstance := db.GetReplicaInstance()
	if replicaInstance == nil {
		return nil, NewError(ModelRuntimeError, "replica instance is nil")
	}
	s.log("replica:", query, args)
	return replicaInstance.QueryContext(s.ctx, query, args...)
}

func (s *Session) Query(query string, args ...interface{}) (rows *sql.Rows, err error) {
	s.txMutex.RLock()
	defer s.txMutex.RUnlock()
	s.log("main:", query, args)
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
	s.log("main:", query, args)
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
		err = s.txBegin()
		if err == nil {
			err = f()
			if err != nil {
				if err := s.txRollback(); err != nil {
					return err
				}
			} else {
				err = s.txCommit()
			}
		}
	}
	return
}

func (s *Session) txBegin() error {
	if s.tx != nil {
		log.Println("session.txBegin: can not begin tx again")
	} else {
		var err error
		if s.tx, err = db.GetInstance().Begin(); err != nil {
			log.Printf("session.txBegin: %s\n", err.Error())
			return err
		}
	}
	return nil
}

func (s *Session) txRollback() error {
	if s.tx != nil {
		if err := s.tx.Rollback(); err != nil {
			log.Printf("session.txRollback: %s\n", err.Error())
			return err
		}
		s.tx = nil
	}
	return nil
}

func (s *Session) txCommit() error {
	if s.tx != nil {
		if err := s.tx.Commit(); err != nil {
			log.Printf("session.txCommit: %s\n", err.Error())
			return err
		}
		s.tx = nil
	}
	return nil
}
