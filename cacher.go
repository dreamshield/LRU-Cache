package cache

import (
	"container/list"
	"crypto/sha1"
	"fmt"
	"sync"
	"time"
)

// LRU cache container
type LRUCacher struct {
	nodeList       *list.List
	nodeIndex      map[string]*list.Element
	store          CacheStore
	mutex          sync.Mutex
	Expired        time.Duration
	MaxElementSize int
	GcInterval     time.Duration
}

func NewLRUCacher(store CacheStore, expired time.Duration, maxElementSize int) *LRUCacher {
	cacher := &LRUCacher{
		store:          store,
		nodeList:       list.New(),
		Expired:        expired,
		GcInterval:     CacheGcInterval,
		MaxElementSize: maxElementSize,
		nodeIndex:      make(map[string]*list.Element),
	}
	cacher.RunGC()
	return cacher
}

// RunGC run once every m.GcInterval
func (m *LRUCacher) RunGC() {
	time.AfterFunc(m.GcInterval, func() {
		m.RunGC()
		m.GC()
	})
}

// GC run once every m.GcInterval to remove all element expired
func (m *LRUCacher) GC() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	var removedNum int = 0

	for e := m.nodeList.Front(); e != nil; {
		next := e.Next()
		if removedNum >= CacheGcMaxRemoved {
			break
		}
		if time.Now().Sub(e.Value.(*node).lastVisit) > m.Expired {
			removedNum++
			tmpNode := e.Value.(*node)
			m.delBean(tmpNode.key)
		}
		e = next
	}
}

// GetBean returns bean according special key from cache
func (m *LRUCacher) GetBean(key string) interface{} {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	storeKey := genNodeKey(key)
	if v, err := m.store.Get(storeKey); err == nil {
		if el, ok := m.nodeIndex[storeKey]; ok {
			lastTime := el.Value.(*node).lastVisit
			// if expired, remove the node and return nil
			if time.Now().Sub(lastTime) > m.Expired {
				m.delBean(storeKey)
				return nil
			}
			m.nodeList.MoveToBack(el)
			el.Value.(*node).lastVisit = time.Now()
		} else {
			el = m.nodeList.PushBack(newNode(storeKey))
			m.nodeIndex[storeKey] = el
		}
		return v
	} else {
		// store bean is not exist, then remove memory's index
		m.delBean(storeKey)
		return nil
	}
}

// clear all the cached data
func (m *LRUCacher) clearBeans() {
	for e := m.nodeList.Front(); e != nil; {
		key := e.Value.(*node).key
		tmp := e
		e = tmp.Next()
		m.nodeList.Remove(tmp)
		m.store.Del(key)
	}
	m.nodeIndex = make(map[string]*list.Element)
}

// clear all the cached data for calling by other package
func (m *LRUCacher) ClearBeans() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.clearBeans()
}

// put data into the memory cache
func (m *LRUCacher) PutBean(key string, obj interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	var el *list.Element
	var ok bool
	storeKey := genNodeKey(key)
	if el, ok = m.nodeIndex[storeKey]; !ok {
		el = m.nodeList.PushBack(newNode(storeKey))
		m.nodeIndex[storeKey] = el
	} else {
		el.Value.(*node).lastVisit = time.Now()
		m.nodeList.MoveToBack(el)
	}

	m.store.Put(storeKey, obj)
	if m.nodeList.Len() > m.MaxElementSize {
		el = m.nodeList.Front()
		m.delBean(el.Value.(*node).key)
	}
}

// delete data by key
func (m *LRUCacher) delBean(key string) {
	if el, ok := m.nodeIndex[key]; ok {
		delete(m.nodeIndex, key)
		m.nodeList.Remove(el)
	}
	m.store.Del(key)
}

// delete data by key for other package
func (m *LRUCacher) DelBean(key string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	storeKey := genNodeKey(key)
	m.delBean(storeKey)
}

// data node record the data key which points the data in the memory cache
// and last visit time
type node struct {
	key       string
	lastVisit time.Time
}

// generate the cache key to store
func genNodeKey(key string) string {
	prefix := "CACHE_PREFIX"
	hash := sha1.New()
	hash.Write([]byte(prefix + key))
	return fmt.Sprintf("%v", hash.Sum(nil))
}

// create the data node for LRU cache
func newNode(key string) *node {
	return &node{key, time.Now()}
}
