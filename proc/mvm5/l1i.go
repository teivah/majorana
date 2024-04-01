package mvm5

type l1i struct {
	boundary [2]int32
	size     int32
}

func newL1I(size int32) l1i {
	return l1i{
		boundary: [2]int32{-1, -1},
		size:     size,
	}
}

func (l *l1i) present(pc int32) bool {
	return pc >= l.boundary[0] && pc <= l.boundary[1]
}

func (l *l1i) fetch(pc int32) {
	l.boundary = [2]int32{pc, pc + l.size}
}
