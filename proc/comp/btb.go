package comp

type BranchTargetBuffer struct {
	buffer []entry
	length int
}

func NewBranchTargetBuffer(length int) *BranchTargetBuffer {
	return &BranchTargetBuffer{length: length}
}

type entry struct {
	pc     int32
	pcDest int32
}

func (b *BranchTargetBuffer) Add(pc, pcDest int32) {
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

func (b *BranchTargetBuffer) Get(pc int32) *int32 {
	for _, e := range b.buffer {
		if e.pc == pc {
			return &e.pcDest
		}
	}
	return nil
}
