package comp

type L1i struct {
	boundary [2]int32
	size     int32
}

func NewL1I(size int32) L1i {
	return L1i{
		boundary: [2]int32{-1, -1},
		size:     size,
	}
}

func (l *L1i) Present(pc int32) bool {
	return pc >= l.boundary[0] && pc <= l.boundary[1]
}

func (l *L1i) Fetch(pc int32) {
	l.boundary = [2]int32{pc, pc + l.size}
}
