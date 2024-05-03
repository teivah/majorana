package cache

import "slices"

type LRUCache[K comparable, V any] struct {
	capacity int
	cache    map[K]V
	order    []K
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		capacity: capacity,
		cache:    make(map[K]V),
		order:    make([]K, 0, capacity),
	}
}

func (l *LRUCache[K, V]) Get(key K) (V, bool) {
	if val, ok := l.cache[key]; ok {
		l.refreshOrder(key)
		return val, true
	}
	var zero V
	return zero, false
}

func (l *LRUCache[K, V]) Find(keys []K) (K, bool) {
	var zero K
	if len(l.order) == 0 {
		return zero, false
	}

	for _, k := range l.order {
		if slices.Contains(keys, k) {
			l.refreshOrder(k)
			return k, true
		}
	}
	return zero, false
}

func (l *LRUCache[K, V]) Put(key K, value V) {
	if _, ok := l.cache[key]; !ok && len(l.cache) == l.capacity {
		delete(l.cache, l.order[0])
		l.order = l.order[1:]
	}
	l.cache[key] = value
	l.refreshOrder(key)
}

func (l *LRUCache[K, V]) refreshOrder(key K) {
	for i := 0; i < len(l.order); i++ {
		if l.order[i] == key {
			l.order = append(l.order[:i], l.order[i+1:]...)
			break
		}
	}
	l.order = append(l.order, key)
}
