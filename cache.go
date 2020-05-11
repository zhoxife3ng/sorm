package sorm

import (
	"container/list"
	"database/sql/driver"
	"fmt"
	"github.com/x554462/go-exception"
	"strings"
	"sync"
)

func (d *Dao) queryCache(indexes ...interface{}) Modeller {
	key := d.buildKey(indexes...)
	if m, err := d.Session().daoModelCache.Get(key); err == nil {
		return m
	}
	return nil
}

func (d *Dao) removeCache(indexes ...interface{}) {
	key := d.buildKey(indexes...)
	d.Session().daoModelCache.Del(key)
}

func (d *Dao) saveCache(model Modeller) {
	key := d.buildKey(model.IndexValues()...)
	d.Session().daoModelCache.Put(key, model)
}

func (d *Dao) buildKey(indexes ...interface{}) string {
	var err error
	buildStr := strings.Builder{}
	buildStr.WriteString(d.tableName)
	for _, v := range indexes {
	assert:
		switch m := v.(type) {
		case int, int64, uint, uint64, int32, int16, int8, uint32, uint16, uint8:
			buildStr.WriteString(fmt.Sprintf("`%d", m))
		case string:
			if m == "" {
				err = exception.New("the index key can not be empty string", ModelRuntimeError)
			}
			buildStr.WriteString("`")
			buildStr.WriteString(m)
		case []byte:
			buildStr.WriteString("`")
			buildStr.Write(m)
		case driver.Valuer:
			if v, err = m.Value(); err == nil {
				goto assert
			}
		default:
			err = exception.New("not support index key", ModelRuntimeError)
		}
		exception.CheckError(err)
	}
	return buildStr.String()
}

// 使用lru算法缓存model
// 只在当前session有效
const maxLength = 200

type element struct {
	listElem *list.Element
	model    Modeller
}

type modelLruCache struct {
	elements map[string]*element
	list     *list.List
	capacity int // 容量
	used     int // 使用量
	locker   sync.RWMutex
}

func newDaoLru(capacity int) *modelLruCache {
	size := maxLength
	if maxLength > capacity {
		size = capacity
	}
	return &modelLruCache{
		elements: make(map[string]*element, size),
		list:     list.New(),
		capacity: capacity,
		used:     0,
	}
}

func (lru *modelLruCache) Clear() {
	lru.elements = make(map[string]*element, lru.used)
	lru.list.Init()
	lru.used = 0
}

func (lru *modelLruCache) Get(key string) (Modeller, error) {
	if lru.used > 0 {
		lru.locker.RLock()
		defer lru.locker.RUnlock()
		if element, ok := lru.elements[key]; ok {
			lru.list.MoveToBack(element.listElem)
			return element.model, nil
		}
	}
	return nil, ModelNotFoundError
}

func (lru *modelLruCache) Put(key string, model Modeller) {

	lru.locker.Lock()
	defer lru.locker.Unlock()

	if elem, ok := lru.elements[key]; ok {
		lru.elements[key] = &element{listElem: elem.listElem, model: model}
		lru.list.MoveToBack(elem.listElem)
		return
	}
	lru.addElement(key, model)
	if lru.used > lru.capacity {
		lru.delListFrontElement()
	}
}

func (lru *modelLruCache) Del(key string) {
	if element, ok := lru.elements[key]; ok {
		lru.list.Remove(element.listElem)
		delete(lru.elements, key)
		lru.used--
	}
}

func (lru *modelLruCache) addElement(key string, model Modeller) {
	lru.used++
	listElem := lru.list.PushBack(key)
	lru.elements[key] = &element{listElem: listElem, model: model}
}

func (lru *modelLruCache) delListFrontElement() {
	frontElem := lru.list.Front()
	key := frontElem.Value.(string)
	lru.Del(key)
}
