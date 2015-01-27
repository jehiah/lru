package lru

import (
	"log"
	"testing"
	"time"
)

func TestLRU(t *testing.T) {
	var total int64
	var removed int64
	var removedKeys []string
	newItem := func(k Key) Value {
		total += 1
		log.Printf("newItem %d", total)
		return total
	}
	removal := func(k Key, v Value) {
		log.Printf("removal %v %v", k, v)
		removed++
		removedKeys = append(removedKeys, k.(string))
	}
	lru := New(newItem, removal, 4, 0)
	for _, k := range []string{"key1", "key2", "key3"} {
		if v, ok := lru.Get(k); ok && v.(int64) != total {
			t.Errorf("%s got %d expected %d", k, v, total)
		}
	}
	for i, k := range []string{"key1", "key2", "key3"} {
		if v, ok := lru.Get(k); ok && v.(int64) != int64(i+1) {
			t.Errorf("%s got %d expected %d", k, v, i+1)
		}
	}

	if removed != 0 {
		t.Errorf("removed = %d", removed)
	}
	lru.Flush()
	if removed != 3 {
		t.Errorf("removed = %d", removed)
	}
	if total != 3 {
		t.Errorf("total = %d", total)
	}
}

func TestExpiry(t *testing.T) {
	var removed int64
	removal := func(k Key, v Value) {
		log.Printf("removal %v %v", k, v)
		removed++
	}
	lru := New(nil, removal, 4, 10*time.Millisecond)
	lru.Set("key", 1)
	_, ok := lru.Get("key")
	if !ok {
		t.Errorf("entry should still be there")
	}
	time.Sleep(10 * time.Millisecond)
	_, ok = lru.Get("key")
	if ok {
		t.Errorf("entry should be expired, but isn't")
	}
	if removed != 1 {
		t.Errorf("removal didn't happen")
	}
}
