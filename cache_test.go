package cache

import (
	"strconv"
	"testing"
	"time"
)

const (
	// EXPIRE_TIME      time.Duration = 60 * time.Minute
	EXPIRE_TIME      time.Duration = 2 * time.Second
	MAX_ELEMENT_SIZE int           = 100
)

func TestCacheInit(t *testing.T) {
	cacher := NewLRUCacher(NewMemoryStore(), EXPIRE_TIME, MAX_ELEMENT_SIZE)

	if cacher == nil {
		t.Error("Initial LRUCacher error")
	}
	t.Logf("cacher pointer = %v", cacher)
}

func TestPutAndGetBean(t *testing.T) {
	var key string
	cacher := NewLRUCacher(NewMemoryStore(), EXPIRE_TIME, MAX_ELEMENT_SIZE)
	if cacher == nil {
		t.Error("Initial LRUCacher error")
	}
	key = "test1"
	obj1 := "test data"
	cacher.PutBean(key, obj1)
	val1 := cacher.GetBean(key)
	if val1 == nil {
		t.Error("put or get methord error")
	}
	t.Logf("val = %v", val1)
	key = "test2"
	obj2 := 123
	cacher.PutBean(key, obj2)
	val2 := cacher.GetBean(key)
	if val2 == nil {
		t.Error("put or get methord error")
	}
	t.Logf("data = %v", val2)

	go func() {
		for k, v := range cacher.nodeIndex {
			t.Logf("key = %v, v = %v", k, v)
			t.Logf("val = %v", v.Value.(*node).key)
			t.Logf("lastVisit = %v", v.Value.(*node).lastVisit)
		}
	}()

	go func() {
		for el := cacher.nodeList.Front(); el != nil; el = el.Next() {
			key = el.Value.(*node).key
			t.Logf("key = %s", string(key))
			v, _ := cacher.store.Get(key)
			// v := cacher.GetBean(key)
			t.Logf("v = %v", v)
		}
	}()
	time.Sleep(2 * time.Second)
}

func TestDeleteBean(t *testing.T) {
	var key string
	cacher := NewLRUCacher(NewMemoryStore(), EXPIRE_TIME, MAX_ELEMENT_SIZE)
	if cacher == nil {
		t.Error("Initial LRUCacher error")
	}
	key = "test1"
	obj1 := "test data"
	cacher.PutBean(key, obj1)
	val1 := cacher.GetBean(key)
	if val1 == nil {
		t.Error("put or get methord error")
	}
	t.Logf("val = %v", val1)
	cacher.DelBean(key)
	val1 = cacher.GetBean(key)
	if val1 != nil {
		tmpKey := genNodeKey(key)
		if val, err := cacher.store.Get(tmpKey); err == nil {
			t.Logf("memval = %v", val)
			t.Error("memory delete methord error")
		}
		t.Logf("val = %v", val1)
		t.Error("cache delete bean methord error")
	}
}

func TestMaxQueElement(t *testing.T) {
	// var key string
	cacher := NewLRUCacher(NewMemoryStore(), EXPIRE_TIME, MAX_ELEMENT_SIZE)
	if cacher == nil {
		t.Error("Initial LRUCacher error")
	}
	var keyMap = make(map[int]string)
	var dataArr []string
	var testLen int = 100
	for i := 0; i < testLen; i++ {
		keyMap[i] = "key-" + strconv.Itoa(i)
		dataArr = append(dataArr, "data-"+strconv.Itoa(i))
	}

	for i := 0; i < 10; i++ {
		key := keyMap[i]
		val := dataArr[i]
		cacher.PutBean(key, val)
	}

	for i := 0; i < 10; i++ {
		key := keyMap[i]
		v := cacher.GetBean(key)
		t.Logf("key=%v,val=%v", key, v)
	}

	for el := cacher.nodeList.Front(); el != nil; el = el.Next() {
		key := el.Value.(*node).key
		v, _ := cacher.store.Get(key)
		t.Logf("v = %v", v)
	}

	for i := 10; i < 20; i++ {
		key := keyMap[i]
		val := dataArr[i]
		cacher.PutBean(key, val)
	}

	for i := 10; i < 20; i++ {
		key := keyMap[i]
		v := cacher.GetBean(key)
		t.Logf("key=%v,val=%v", key, v)
	}

	for el := cacher.nodeList.Front(); el != nil; el = el.Next() {
		key := el.Value.(*node).key
		v, _ := cacher.store.Get(key)
		t.Logf("v = %v", v)
	}

	for i := 0; i < 10; i++ {
		key := keyMap[i]
		v := cacher.GetBean(key)
		t.Logf("key=%v,val=%v", key, v)
	}
}

func TestCacheGC(t *testing.T) {
	cacher := NewLRUCacher(NewMemoryStore(), EXPIRE_TIME, MAX_ELEMENT_SIZE)
	if cacher == nil {
		t.Error("Initial LRUCacher error")
	}
	var keyMap = make(map[int]string)
	var dataArr []string
	var testLen int = 100
	for i := 0; i < testLen; i++ {
		keyMap[i] = "key-" + strconv.Itoa(i)
		dataArr = append(dataArr, "data-"+strconv.Itoa(i))
	}

	for i := 0; i < 10; i++ {
		key := keyMap[i]
		val := dataArr[i]
		cacher.PutBean(key, val)
	}

	t.Log("----------before gc--------")

	for i := 0; i < 10; i++ {
		key := keyMap[i]
		val := cacher.GetBean(key)
		if val == nil {
			t.Error("get bean from cache error")
		}
		t.Logf("val=%v", val)
	}

	for el := cacher.nodeList.Front(); el != nil; el = el.Next() {
		key := el.Value.(*node).key
		v, _ := cacher.store.Get(key)
		t.Logf("v = %v", v)
	}

	time.Sleep(4 * time.Second)

	t.Log("---------after gc---------")

	for el := cacher.nodeList.Front(); el != nil; el = el.Next() {
		key := el.Value.(*node).key
		v, _ := cacher.store.Get(key)
		t.Logf("v = %v", v)
	}

	t.Log("add data twice")

	for i := 10; i < 20; i++ {
		key := keyMap[i]
		val := dataArr[i]
		cacher.PutBean(key, val)
	}

	time.Sleep(1 * time.Second)

	t.Log("-------before gc-----------")
	for el := cacher.nodeList.Front(); el != nil; el = el.Next() {
		key := el.Value.(*node).key
		v, _ := cacher.store.Get(key)
		t.Logf("v = %v", v)
	}

	time.Sleep(2 * time.Second)

	t.Log("--------after gc----------")

	for el := cacher.nodeList.Front(); el != nil; el = el.Next() {
		key := el.Value.(*node).key
		v, _ := cacher.store.Get(key)
		t.Logf("v = %v", v)
	}
}

func TestGCMaxRemovedNum(t *testing.T) {
	cacher := NewLRUCacher(NewMemoryStore(), EXPIRE_TIME, MAX_ELEMENT_SIZE)
	if cacher == nil {
		t.Error("Initial LRUCacher error")
	}
	var keyMap = make(map[int]string)
	var dataArr []string
	var testLen int = 100
	for i := 0; i < testLen; i++ {
		keyMap[i] = "key-" + strconv.Itoa(i)
		dataArr = append(dataArr, "data-"+strconv.Itoa(i))
	}

	for i := 0; i < 10; i++ {
		key := keyMap[i]
		val := dataArr[i]
		cacher.PutBean(key, val)
	}

	t.Log("----------before gc--------")

	for el := cacher.nodeList.Front(); el != nil; el = el.Next() {
		key := el.Value.(*node).key
		v, _ := cacher.store.Get(key)
		t.Logf("v = %v", v)
	}

	time.Sleep(3 * time.Second)

	t.Log("---------after gc---------")

	for el := cacher.nodeList.Front(); el != nil; el = el.Next() {
		key := el.Value.(*node).key
		v, _ := cacher.store.Get(key)
		t.Logf("v = %v", v)
	}

	for i := 0; i < 10; i++ {
		key := keyMap[i]
		val := cacher.GetBean(key)
		if val != nil {
			t.Error("gc removed number control error")
			t.Logf("val=%v", val)
		}
	}

	for el := cacher.nodeList.Front(); el != nil; el = el.Next() {
		key := el.Value.(*node).key
		v, _ := cacher.store.Get(key)
		t.Logf("v = %v", v)
	}
}

func TestGCClear(t *testing.T) {
	cacher := NewLRUCacher(NewMemoryStore(), EXPIRE_TIME, MAX_ELEMENT_SIZE)
	if cacher == nil {
		t.Error("Initial LRUCacher error")
	}
	var keyMap = make(map[int]string)
	var dataArr []string
	var testLen int = 20
	for i := 0; i < testLen; i++ {
		keyMap[i] = "key-" + strconv.Itoa(i)
		dataArr = append(dataArr, "data-"+strconv.Itoa(i))
	}

	for i := 0; i < testLen; i++ {
		key := keyMap[i]
		val := dataArr[i]
		cacher.PutBean(key, val)
	}

	t.Log("before clear")

	for el := cacher.nodeList.Front(); el != nil; el = el.Next() {
		key := el.Value.(*node).key
		v, _ := cacher.store.Get(key)
		t.Logf("key = %v, v = %v", key, v)
	}

	cacher.ClearBeans()

	t.Log("after clear")

	for el := cacher.nodeList.Front(); el != nil; el = el.Next() {
		key := el.Value.(*node).key
		if val, err := cacher.store.Get(key); err == nil {
			t.Errorf("clear cache error, val=%v, key=%v", val, key)
		}
	}

}
