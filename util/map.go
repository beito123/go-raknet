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
	"sort"
	"strconv"

	"github.com/orcaman/concurrent-map"
)

func NewIntMap() IntMap {
	return IntMap{
		Map: cmap.New(),
	}
}

// IntMap is a simple int map
type IntMap struct {
	Map cmap.ConcurrentMap
}

func (m *IntMap) IsEmpty() bool {
	return m.Map.IsEmpty()
}

func (m *IntMap) Size() int {
	return m.Map.Count()
}

func (m *IntMap) Has(key int) bool {
	return m.Map.Has(strconv.Itoa(key))
}

func (m *IntMap) Get(key int) (interface{}, bool) {
	return m.Map.Get(strconv.Itoa(key))
}

func (m *IntMap) Set(key int, value interface{}) {
	m.Map.Set(strconv.Itoa(key), value)
}

func (m *IntMap) Remove(key int) {
	m.Remove(key)
}

func (m *IntMap) Clear() {
	m.Map = cmap.New()
}

func (m *IntMap) Poll() (val interface{}, ok bool) {
	keys := m.keys()
	if len(keys) == 0 {
		return nil, false
	}

	val, ok = m.Map.Get(keys[0])
	if !ok {
		return nil, false
	}

	m.Map.Remove(keys[0])

	return val, true
}

func (m *IntMap) Range(f func(key int, value interface{}) bool) error {
	items := m.Map.Items()
	for _, k := range m.keys() {
		item, ok := items[k]
		if !ok {
			continue
		}

		n, err := strconv.Atoi(k)
		if err != nil {
			continue
		}

		if !f(n, item) {
			break
		}
	}

	return nil
}

func (m *IntMap) keys() []string {
	keys := Strings(m.Map.Keys())

	sort.Sort(keys)

	return keys
}
