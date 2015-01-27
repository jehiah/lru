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

// Counter with eviction for least-recently used (LRU) items.

package lru

// a LRU counter that calls a function when an item is removed
type LRUCounter struct {
	lru *LRU
}

// Create a new LRU cache for function f with the desired capacity.
func NewLRUCounter(removalFunc func(interface{}, int64), capacity int) *LRUCounter {
	r := func(key Key, value Value) {
		vv := value.(int64)
		removalFunc(key, vv)
	}
	l := New(nil, r, capacity, 0)
	return &LRUCounter{l}
}

// Fetch value for key in the cache, updating it's LRU position
func (c *LRUCounter) Get(key interface{}) (value int64, ok bool) {
	v, ok := c.lru.Get(key)
	if ok {
		value = v.(int64)
	}
	return value, ok
}

// Number of items currently in the cache.
func (c *LRUCounter) Len() int {
	return c.lru.Len()
}

func (c *LRUCounter) Capacity() int {
	return c.lru.Capacity()
}

func (c *LRUCounter) Flush() {
	c.lru.Flush()
}

func (c *LRUCounter) Incr(key interface{}, value int64) {
	if vv, ok := c.Get(key); ok {
		value += vv
	}
	c.lru.Set(key, value)
}
