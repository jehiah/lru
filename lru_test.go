package lru

import (
	"testing"
	"log"
)

func TestLRU(t *testing.T) {
	var total int64
	var removed int64
	var removedKeys []string
	newItem := func(k interface{}) interface{} {
		total += 1
		log.Printf("newItem %d", total)
		return total
	}
	removal := func(k, v interface{}) {
		log.Printf("removal %v %v", k, v)
		removed++
		removedKeys = append(removedKeys, k.(string))
	}
	lru := New(newItem, removal, 4)
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
