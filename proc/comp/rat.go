package comp

// RAT is the register allocation table
type RAT[K comparable, V any] struct {
	length int
	values map[K][]V
	idx    map[K]int
}

func NewRAT[K comparable, V any](length int) *RAT[K, V] {
	return &RAT[K, V]{
		length: length,
		values: make(map[K][]V),
		idx:    make(map[K]int),
	}
}

func (r *RAT[K, V]) Read(k K) (V, bool) {
	var zero V

	idx, exists := r.idx[k]
	if !exists {
		return zero, false
	}
	return r.values[k][idx], true
}

func (r *RAT[K, V]) Find(k K, predicate func(V) bool) (V, bool) {
	var zero V

	idx, exists := r.idx[k]
	if !exists {
		return zero, false
	}

	for i := idx; i >= 0; i-- {
		v := r.values[k][idx]
		if predicate(v) {
			return v, true
		}
	}

	for i := r.length - 1; i > idx; i-- {
		v := r.values[k][idx]
		if predicate(v) {
			return v, true
		}
	}

	return zero, false
}

func (r *RAT[K, V]) Write(k K, value V) {
	idx, exists := r.idx[k]
	if !exists {
		idx = 0
		r.values[k] = make([]V, r.length)
	} else {
		idx = (idx + 1) % r.length
	}

	r.idx[k] = idx
	r.values[k][idx] = value
}

func (r *RAT[K, V]) Values() map[K]V {
	m := make(map[K]V)
	for k, v := range r.idx {
		m[k] = r.values[k][v]
	}
	return m
}

func (r *RAT[K, V]) FindValues(predicate func(V) bool) map[K]V {
	m := make(map[K]V)
	for k, v := range r.idx {
		found := false
		for i := v; i >= 0; i-- {
			if predicate(r.values[k][i]) {
				m[k] = r.values[k][i]
				found = true
				break
			}
		}
		if found {
			continue
		}
		for i := r.length - 1; i > v; i-- {
			if predicate(r.values[k][i]) {
				m[k] = r.values[k][i]
				break
			}
		}
	}
	return m
}
