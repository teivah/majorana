package mvp5_1

type branchTargetBuffer struct {
	buffer []entry
	length int
}

func newBranchTargetBuffer(length int) *branchTargetBuffer {
	return &branchTargetBuffer{length: length}
}

type entry struct {
	pc     int32
	pcDest int32
}

func (b *branchTargetBuffer) add(pc, pcDest int32) {
	for i := 0; i < len(b.buffer); i++ {
		e := b.buffer[i]
		if e.pc == pc {
			e.pcDest = pcDest
			b.buffer[i] = e
			return
		}
	}

	e := entry{
		pc:     pc,
		pcDest: pcDest,
	}
	if len(b.buffer) != b.length {
		b.buffer = append(b.buffer, e)
	} else {
		b.buffer = append(b.buffer[1:], e)
	}
}

func (b *branchTargetBuffer) get(pc int32) (int32, bool) {
	for _, e := range b.buffer {
		if e.pc == pc {
			return e.pcDest, true
		}
	}
	return 0, false
}
