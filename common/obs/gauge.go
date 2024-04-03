package obs

type Gauge struct {
	sum   int
	count int
}

func (g *Gauge) Push(v int) {
	g.sum += v
	g.count++
}

func (g *Gauge) Stats() float64 {
	return float64(g.sum) / float64(g.count)
}
