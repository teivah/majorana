package mvm5

type gauge struct {
	sum   int
	count int
}

func (g *gauge) push(v int) {
	g.sum += v
	g.count++
}

func (g *gauge) stats() float64 {
	return float64(g.sum) / float64(g.count)
}
