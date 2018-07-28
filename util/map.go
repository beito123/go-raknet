package util

/*
 * go-raknet
 *
 * Copyright (c) 2018 beito
 *
 * This software is released under the MIT License.
 * http://opensource.org/licenses/mit-license.php
 */

import (
	"math"
	"sync"

	"github.com/secnot/orderedmap"
)

func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		Map: orderedmap.NewOrderedMap(),
	}
}

// OrderedMap is an ordered map supported concurrent acceses
// It's a wrapper of github.com/secnot/orderedmap. Thank you!
type OrderedMap struct {
	Map *orderedmap.OrderedMap

	sync.RWMutex
}

func (m *OrderedMap) Clear() {
	m.Lock()
	m.Map = orderedmap.NewOrderedMap()
	m.Unlock()
}

func (m *OrderedMap) Len() (l int) {
	m.RLock()
	l = m.Map.Len()
	m.RUnlock()

	return l
}

func (m *OrderedMap) Exist(key interface{}) (ok bool) {
	_, ok = m.Get(key)
	return ok
}

func (m *OrderedMap) Get(key interface{}) (val interface{}, ok bool) {
	m.RLock()
	val, ok = m.Map.Get(key)
	m.RUnlock()

	return val, ok
}

func (m *OrderedMap) Set(key interface{}, val interface{}) {
	m.Lock()
	m.Map.Set(key, val)
	m.Unlock()
}

func (m *OrderedMap) first() (key interface{}, val interface{}, ok bool) {
	m.RLock()
	key, val, ok = m.Map.GetFirst()
	m.RUnlock()

	return key, val, ok
}

func (m *OrderedMap) last() (key interface{}, val interface{}, ok bool) {
	m.RLock()
	key, val, ok = m.Map.GetLast()
	m.RUnlock()

	return key, val, ok
}

func (m *OrderedMap) FirstKey() (key interface{}, ok bool) {
	key, _, ok = m.first()

	return key, ok
}

func (m *OrderedMap) LastKey() (key interface{}, ok bool) {
	key, _, ok = m.last()

	return key, ok
}

func (m *OrderedMap) First() (val interface{}, ok bool) {
	_, val, ok = m.first()

	return val, ok
}

func (m *OrderedMap) Last() (val interface{}, ok bool) {
	_, val, ok = m.last()

	return val, ok
}

func (m *OrderedMap) Remove(key interface{}) {
	m.Lock()
	m.Map.Delete(key)
	m.Unlock()
}

func (m *OrderedMap) Pop() (val interface{}, ok bool) {
	key, val, ok := m.Map.GetLast()
	if !ok {
		return nil, false
	}

	m.Remove(key)

	return val, ok
}

func (m *OrderedMap) Shift() (val interface{}, ok bool) {
	key, val, ok := m.Map.GetFirst()
	if !ok {
		return nil, false
	}

	m.Remove(key)

	return val, ok
}

func (m *OrderedMap) Range(f func(key interface{}, value interface{}) bool) {
	iter := m.Map.Iter()
	for {
		key, val, ok := iter.Next()
		if !ok {
			break
		}

		if !f(key, val) {
			break
		}
	}
}

func (m *OrderedMap) RangeAndRemove(f func(key interface{}, value interface{}) bool) {
	iter := m.Map.Iter()
	for {
		key, val, ok := iter.Next()
		if !ok {
			break
		}

		m.Remove(key)

		if !f(key, val) {
			break
		}
	}
}

func NewQueue() *Queue {
	return &Queue{
		Map: NewOrderedMap(),
	}
}

type Queue struct {
	Map *OrderedMap

	off int
}

func (q *Queue) Clear() {
	q.off = 0

	q.Clear()
}

func (q *Queue) IsEmpty() bool {
	return q.Size() == 0
}

func (q *Queue) Size() int {
	return q.Size()
}

func (q *Queue) Remove() {
	key, ok := q.Map.LastKey()
	if !ok {
		return
	}

	q.Map.Remove(key)
}

func (q *Queue) bump() int {
	return (q.off % math.MaxInt32) + 1
}

func (q *Queue) Add(val interface{}) {
	q.Map.Set(q.bump(), val)
}

func (q *Queue) Peek() (interface{}, bool) {
	return q.Map.Last()
}

func (q *Queue) Poll() (interface{}, bool) {
	return q.Map.Pop()
}

func (q *Queue) Range(f func(val interface{}) bool) {
	q.Map.Range(func(key interface{}, val interface{}) bool {
		return f(val)
	})
}

func (q *Queue) RangeAndRemove(f func(val interface{}) bool) {
	q.Map.RangeAndRemove(func(key interface{}, val interface{}) bool {
		return f(val)
	})
}
