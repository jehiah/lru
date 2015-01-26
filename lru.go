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

type AddFunc func(interface{}) interface{}
type RemovalFunc func(interface{}, interface{})

// Cache for function Func.
type LRU struct {
	addFunc     AddFunc
	removalFunc RemovalFunc
	index       map[interface{}]int // index of key in queue
	queue       list
}

// Create a new LRU cache with the desired capacity and optional functions to add new items, or
// notify on removal
func New(a AddFunc, r RemovalFunc, capacity int) *LRU {
	if capacity < 1 {
		panic("capacity < 1")
	}
	c := &LRU{addFunc: a, removalFunc: r, index: make(map[interface{}]int)}
	c.queue.init(capacity)
	return c
}

// Fetch value for key in the cache, calling AddFunc to compute it if necessary.
// This updates the values position in the LRU cache
func (c *LRU) Get(key interface{}) (value interface{}, ok bool) {
	i, ok := c.index[key]
	if ok {
		value = c.queue.valueAt(i)
		c.queue.moveToFront(i)
	} else {
		if c.addFunc != nil {
			value = c.addFunc(key)
			c.insert(key, value)
			ok = true
		}
	}
	return value, ok
}

// Set a new entry in the LRU cache
func (c *LRU) Set(key, value interface{}) {
	i, ok := c.index[key]
	if ok {
		c.queue.links[i].value = value
		c.queue.moveToFront(i)
	} else {
		c.insert(key, value)
	}
}

// Number of items currently in the LRU cache.
func (c *LRU) Len() int {
	return len(c.queue.links)
}

func (c *LRU) Capacity() int {
	return cap(c.queue.links)
}

// Iterate over the cache in LRU order. Useful for debugging.
func (c *LRU) Iter(keys chan interface{}, values chan interface{}) {
	for i := c.queue.tail; i != -1; {
		n := c.queue.links[i]
		keys <- n.key
		values <- n.value
		i = n.next
	}
	close(keys)
	close(values)
}

// Flush all entries calling RemovalFunc as needed
func (c *LRU) Flush() {
	if c.removalFunc != nil {
		for i := c.queue.tail; i != -1; {
			n := c.queue.links[i]
			c.removalFunc(n.key, n.value)
			i = n.next
		}
	}
	c.index = make(map[interface{}]int)
	c.queue.init(cap(c.queue.links))
}

func (c *LRU) insert(key interface{}, value interface{}) {
	var i int
	q := &c.queue
	if q.full() {
		// evict least recently used item
		var k, v interface{}
		i, k, v = q.popTail()
		if c.removalFunc != nil {
			c.removalFunc(k, v)
		}
		delete(c.index, k)
	} else {
		i = q.grow()
	}
	q.putFront(key, value, i)
	c.index[key] = i
}

// Doubly linked list containing key/value pairs.
type list struct {
	front, tail, empty int
	links              []link
}

type link struct {
	key, value interface{}
	prev, next int
}

// Initialize l with capacity c.
func (l *list) init(c int) {
	l.front = -1
	l.tail = -1
	l.links = make([]link, 0, c)
}

func (l *list) full() bool {
	return len(l.links) == cap(l.links)
}

// Grow list by one element and return its index.
func (l *list) grow() (i int) {
	i = len(l.links)
	l.links = l.links[:i+1]
	return
}

// Make the node at position i the front of the list.
// Precondition: the list is not empty.
func (l *list) moveToFront(i int) {
	nf := &l.links[i]
	of := &l.links[l.front]

	nf.prev = l.front
	of.next = i
	l.front = i
}

// Pop the tail off the list and return its index and its key.
// Precondition: the list is full.
func (l *list) popTail() (i int, key, value interface{}) {
	i = l.tail
	t := &l.links[i]
	key = t.key
	value = t.value
	l.links[t.next].prev = -1
	l.tail = t.next
	return
}

// Put (key, value) in position i and make it the front of the list.
func (l *list) putFront(key, value interface{}, i int) {
	f := &l.links[i]
	f.key = key
	f.value = value
	f.prev = l.front
	f.next = -1

	if l.tail == -1 {
		l.tail = i
	} else {
		l.links[l.front].next = i
	}
	l.front = i
}

func (l *list) valueAt(i int) interface{} {
	return l.links[i].value
}
