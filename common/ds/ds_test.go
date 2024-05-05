package ds

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type key struct {
	a int
	b int
}

func TestStableMapIteration(t *testing.T) {
	for i := 0; i < 1_000; i++ {
		m := map[key]string{
			key{2, 0}: "3",
			key{1, 4}: "2",
			key{1, 3}: "1",
		}
		var res []string
		less := []func(key) int{
			func(k key) int { return k.a },
			func(k key) int { return k.b },
		}
		for e := range StableMapIteration(m, less) {
			res = append(res, e.V)
		}
		assert.Equal(t, []string{"1", "2", "3"}, res)
	}
}
