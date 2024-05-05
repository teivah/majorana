package ds

import (
	"cmp"
	"sort"
)

type Elem[K comparable, V any] struct {
	K K
	V V
}

func StableMapIteration[K comparable, V any, O cmp.Ordered](m map[K]V, comparables []func(K) O) <-chan Elem[K, V] {
	ch := make(chan Elem[K, V])

	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		a := keys[i]
		b := keys[j]
		for _, c := range comparables {
			v1 := c(a)
			v2 := c(b)
			if v1 < v2 {
				return true
			}
			if v2 < v1 {
				return false
			}
		}
		return true
	})

	go func() {
		for _, k := range keys {
			v := m[k]
			ch <- Elem[K, V]{k, v}
		}
		close(ch)
	}()
	return ch
}
