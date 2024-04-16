package comp

// RAT is the register allocation table
type RAT[K comparable, V any] struct {
	length int
	toIdx  map[K]int
	toReg  map[int]K
	data   []V
	next   int
}

func NewRAT[K comparable, V any](length int) *RAT[K, V] {
	return &RAT[K, V]{
		length: length,
		toIdx:  make(map[K]int),
		toReg:  make(map[int]K),
		data:   make([]V, length),
	}
}

func (r *RAT[K, V]) Read(k K) (V, bool) {
	var zero V
	v, exists := r.toIdx[k]
	if !exists {
		return zero, false
	}
	return r.data[v], true
}

func (r *RAT[K, V]) Write(k K, value V) {
	reg, exists := r.toReg[r.next]
	if exists {
		if r.toIdx[reg] == r.next {
			delete(r.toIdx, reg)
		}
	}

	r.toIdx[k] = r.next
	r.toReg[r.next] = k
	r.data[r.next] = value

	r.next++
	if r.next == r.length {
		r.next = 0
	}
}

func (r *RAT[K, V]) Values() map[K]V {
	m := make(map[K]V)
	for k, v := range r.toIdx {
		m[k] = r.data[v]
	}
	return m
}
