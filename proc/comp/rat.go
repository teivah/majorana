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
