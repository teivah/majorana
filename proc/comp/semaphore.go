package comp

type Sem struct {
	read  int
	write int
}

func (s *Sem) RLock() bool {
	if s.write > 0 {
		return false
	}
	s.read++
	return true
}

func (s *Sem) RUnlock() {
	s.read--
	if s.read < 0 {
		panic("read is negative")
	}
}

func (s *Sem) Lock() bool {
	if s.write > 0 || s.read > 0 {
		return false
	}
	s.write++
	return true
}

func (s *Sem) Unlock() {
	s.write--
	if s.write < 0 {
		panic("write is negative")
	}
}
