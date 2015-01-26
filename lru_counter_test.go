package lru

import (
	"testing"
)

func TestLRUCounter(t *testing.T) {
	var removed int64
	var total int64
	var removedKeys []string

	removalFunc := func(k interface{}, v int64) {
		removed += 1
		total += v
		removedKeys = append(removedKeys, k.(string))
	}
	lru := NewLRUCounter(removalFunc, 4)
	lru.Incr("key1", 1)
	lru.Incr("key2", 1)
	lru.Incr("key2", 1)
	lru.Incr("key3", 1)
	lru.Incr("key3", 1)
	lru.Incr("key4", 1)
	lru.Incr("key5", 1)

	if removed != 1 {
		t.Errorf("removed = %d", removed)
	}
	lru.Flush()
	if removed != 5 {
		t.Errorf("removed = %d", removed)
	}
	if total != 7 {
		t.Errorf("total = %d", total)
	}
}
