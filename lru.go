// Copyright 2013 Lars Buitinck
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

// LRU Cache for arbitrary data with least-recently used (LRU) eviction
// strategy.

package lru

import (
	"container/list"
	"sync"
	"time"
)

type Key interface{}
type Value interface{}
type AddFunc func(Key) Value
type RemovalFunc func(Key, Value)

// container for user data
type entry struct {
	Key        Key
	Value      Value
	lastUpdate time.Time
}

// Cache for function Func.
type LRU struct {
	mu          sync.Mutex
	addFunc     AddFunc
	removalFunc RemovalFunc
	list        *list.List
	table       map[Key]*list.Element
	// how many entries we are lmiting to
	capacity int
	ttl      time.Duration // how long a value is considered good for (0 to disable)
}

// Create a new LRU cache with the desired capacity and optional functions to fetch new items, or
// notify on removal. If a TTL is set, entries will only be considered valid for the TTL duration
func New(a AddFunc, r RemovalFunc, capacity int, ttl time.Duration) *LRU {
	if capacity < 1 {
		panic("capacity < 1")
	}

	return &LRU{
		addFunc:     a,
		removalFunc: r,
		list:        list.New(),
		table:       make(map[Key]*list.Element),
		capacity:    capacity,
		ttl:         ttl,
	}
}

// Fetch value for key in the cache, calling AddFunc to compute it if necessary.
// This updates the values position in the LRU cache
func (lru *LRU) Get(key Key) (v Value, ok bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	element := lru.table[key]
	if element == nil {
		if lru.addFunc != nil {
			v := lru.addFunc(key)
			lru.addNew(key, v)
			return v, true
		}
		return nil, false
	}
	e := element.Value.(*entry)
	if lru.ttl > 0 {
		delta := time.Now().Sub(e.lastUpdate)
		if delta > lru.ttl {
			if lru.removalFunc != nil {
				lru.removalFunc(e.Key, e.Value)
			}
			lru.list.Remove(element)
			delete(lru.table, key)

			// now we also need to conditionally fill this
			if lru.addFunc != nil {
				v := lru.addFunc(key)
				lru.addNew(key, v)
				return v, true
			}
			return nil, false
		}
	}
	lru.list.MoveToFront(element)
	return element.Value.(*entry).Value, true
}

// Set a new entry in the LRU cache
func (lru *LRU) Set(key Key, value Value) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if element := lru.table[key]; element != nil {
		lru.updateInplace(element, value)
	} else {
		lru.addNew(key, value)
	}
}

func (lru *LRU) Delete(key Key) bool {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	element := lru.table[key]
	if element == nil {
		return false
	}

	if lru.removalFunc != nil {
		n := element.Value.(*entry)
		lru.removalFunc(n.Key, n.Value)
	}

	lru.list.Remove(element)
	delete(lru.table, key)
	return true
}

// Number of items currently in the LRU cache.
func (lru *LRU) Len() int {
	return lru.list.Len()
}

func (lru *LRU) Capacity() int {
	return lru.capacity
}

// Iterate over the cache in LRU order. Useful for debugging.
func (lru *LRU) Iter(keys chan Key, values chan Value) {
	for e := lru.list.Front(); e != nil; e = e.Next() {
		keys <- e.Value.(*entry).Key
		values <- e.Value.(*entry).Value
	}
	close(keys)
	close(values)
}

// Flush all entries calling RemovalFunc as needed
func (lru *LRU) Flush() {
	if lru.removalFunc != nil {
		for e := lru.list.Front(); e != nil; e = e.Next() {
			n := e.Value.(*entry)
			lru.removalFunc(n.Key, n.Value)
		}
	}
	lru.list.Init()
	lru.table = make(map[Key]*list.Element)
}

func (lru *LRU) updateInplace(element *list.Element, value Value) {
	e := element.Value.(*entry)
	e.Value = value
	if lru.ttl > 0 {
		e.lastUpdate = time.Now()
	}
	lru.list.MoveToFront(element)
}

func (lru *LRU) addNew(key Key, value Value) {
	e := &entry{Key: key, Value: value}
	if lru.ttl > 0 {
		e.lastUpdate = time.Now()
	}
	element := lru.list.PushFront(e)
	lru.table[key] = element
	lru.checkCapacity()
}

func (lru *LRU) checkCapacity() {
	// Partially duplicated from Delete
	for lru.list.Len() > lru.capacity {
		delElem := lru.list.Back()
		delValue := delElem.Value.(*entry)
		lru.list.Remove(delElem)
		delete(lru.table, delValue.Key)
		if lru.removalFunc != nil {
			lru.removalFunc(delValue.Key, delValue.Value)
		}
	}
}
