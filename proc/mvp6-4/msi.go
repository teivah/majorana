package mvp6_4

import (
	"github.com/teivah/majorana/proc/comp"
)

type msiState = int32

const (
	invalid msiState = iota // Default has to be invalid
	shared
	modified
)

type request struct {
	request     requestType
	alignedAddr int32
	doneFlag    bool
	callback    func()
}

func (r *request) isDone() bool {
	return r.doneFlag
}

func (r *request) done() {
	r.doneFlag = true
	r.callback()
}

type requestType = int32

const (
	// Make sure a zero value isn't confused with an element
	evict requestType = iota + 1
	writeBack
)

type msiResponse struct {
	pendings []*request
	// Mutually exclusive
	wait            bool
	fetchFromMemory bool
	readFromL1      bool
	writeToL1       bool
}

type msiEntry struct {
	id          int
	alignedAddr int32
}

type msi struct {
	pendings map[int32]*comp.Sem
	states   map[msiEntry]msiState
	commands map[msiEntry]*request
}

func newMSI() *msi {
	return &msi{
		pendings: make(map[int32]*comp.Sem),
		states:   make(map[msiEntry]msiState),
		commands: make(map[msiEntry]*request),
	}
}

var noop = func() {}

func (m *msi) requests(id int) []*request {
	var requests []*request
	for e, req := range m.commands {
		if e.id != id {
			continue
		}
		requests = append(requests, req)
	}
	return requests
}

// pendings: pre
// msiResponse: action
// callback: post
func (m *msi) rLock(id int, addrs []int32) (msiResponse, func(), *comp.Sem) {
	alignedAddr := getAlignedMemoryAddress(addrs)
	state := m.getState(id, addrs)
	var pendings []*request
	switch state {
	case invalid:
		if !m.getSem(addrs).RLock() {
			return msiResponse{wait: true}, noop, nil
		}

		// Read request
		for e, state := range m.states {
			if id == e.id {
				continue
			}
			if alignedAddr != e.alignedAddr {
				continue
			}
			switch state {
			case invalid:
				// Nothing
			case modified:
				entry := msiEntry{
					id:          e.id,
					alignedAddr: alignedAddr,
				}
				r := &request{
					request:     writeBack,
					alignedAddr: alignedAddr,
					callback: func() {
						m.states[entry] = invalid
						//fmt.Println("transition", e.id, "invalid")
						delete(m.commands, entry)
					},
				}
				//fmt.Println("read-request", id, e.id)
				if v, exists := m.commands[entry]; exists {
					if r.request != v.request || r.alignedAddr != v.alignedAddr {
						panic("invalid state")
					}
					pendings = append(pendings, v)
					m.commands[entry] = v
				} else {
					pendings = append(pendings, r)
					m.commands[entry] = r
				}
			case shared:
				// Nothing
			}
		}

		return msiResponse{
				fetchFromMemory: true,
				pendings:        pendings,
			}, func() {
				m.setState(id, addrs, shared)
				m.getSem(addrs).RUnlock()
			}, m.getSem(addrs)
	case modified:
		if !m.getSem(addrs).Lock() {
			return msiResponse{wait: true}, noop, nil
		}
		return msiResponse{readFromL1: true}, func() {
			m.getSem(addrs).Unlock()
		}, m.getSem(addrs)
	case shared:
		if !m.getSem(addrs).RLock() {
			return msiResponse{wait: true}, noop, nil
		}
		return msiResponse{readFromL1: true}, func() {
			m.getSem(addrs).RUnlock()
		}, m.getSem(addrs)
	default:
		panic(state)
	}
}

// pendings: pre
// msiResponse: action
// callback: post
func (m *msi) lock(id int, addrs []int32) (msiResponse, func(), *comp.Sem) {
	alignedAddr := getAlignedMemoryAddress(addrs)
	state := m.getState(id, addrs)
	var pendings []*request
	switch state {
	case invalid:
		if !m.getSem(addrs).Lock() {
			return msiResponse{wait: true}, noop, nil
		}

		// Write request
		for e, state := range m.states {
			if id == e.id {
				continue
			}
			if alignedAddr != e.alignedAddr {
				continue
			}
			switch state {
			case invalid:
				// Nothing
			case modified:
				entry := msiEntry{
					id:          e.id,
					alignedAddr: alignedAddr,
				}
				r := &request{
					request:     writeBack,
					alignedAddr: alignedAddr,
					callback: func() {
						m.states[entry] = invalid
						//fmt.Println("transition", e.id, "invalid")
						delete(m.commands, entry)
					},
				}
				//fmt.Println("write-request", id, e.id)
				if v, exists := m.commands[entry]; exists {
					if r.request != v.request || r.alignedAddr != v.alignedAddr {
						panic("invalid state")
					}
					pendings = append(pendings, v)
					m.commands[entry] = v
				} else {
					pendings = append(pendings, r)
					m.commands[entry] = r
				}
			case shared:
				entry := msiEntry{
					id:          e.id,
					alignedAddr: alignedAddr,
				}
				r := &request{
					request:     evict,
					alignedAddr: alignedAddr,
					callback: func() {
						m.states[entry] = invalid
						//fmt.Println("transition", e.id, "invalid")
						delete(m.commands, entry)
					},
				}
				//fmt.Println("invalid-request", id, e.id)
				if v, exists := m.commands[entry]; exists {
					if r.request != v.request || r.alignedAddr != v.alignedAddr {
						panic("invalid state")
					}
					pendings = append(pendings, v)
					m.commands[entry] = v
				} else {
					pendings = append(pendings, r)
					m.commands[entry] = r
				}
			}
		}

		return msiResponse{
				fetchFromMemory: true,
				pendings:        pendings,
			}, func() {
				m.setState(id, addrs, modified)
				m.getSem(addrs).Unlock()
			}, m.getSem(addrs)
	case modified:
		if !m.getSem(addrs).Lock() {
			return msiResponse{wait: true}, noop, nil
		}
		return msiResponse{writeToL1: true}, func() {
			m.getSem(addrs).Unlock()
		}, m.getSem(addrs)
	case shared:
		if !m.getSem(addrs).Lock() {
			return msiResponse{wait: true}, noop, nil
		}

		// Invalidation request
		for e, state := range m.states {
			if id == e.id {
				continue
			}
			if alignedAddr != e.alignedAddr {
				continue
			}
			switch state {
			case invalid:
				// Nothing
			case modified:
				entry := msiEntry{
					id:          e.id,
					alignedAddr: alignedAddr,
				}
				r := &request{
					request:     writeBack,
					alignedAddr: alignedAddr,
					callback: func() {
						m.states[entry] = invalid
						//fmt.Println("transition", e.id, "invalid")
						delete(m.commands, entry)
					},
				}
				pendings = append(pendings, r)
				m.commands[entry] = r
			case shared:
				entry := msiEntry{
					id:          e.id,
					alignedAddr: alignedAddr,
				}
				r := &request{
					request:     evict,
					alignedAddr: alignedAddr,
					callback: func() {
						m.states[entry] = invalid
						//fmt.Println("transition", e.id, "invalid")
						delete(m.commands, entry)
					},
				}
				//fmt.Println("invalid-request", id, e.id)
				pendings = append(pendings, r)
				m.commands[entry] = r
			}
		}

		return msiResponse{
				writeToL1: true,
				pendings:  pendings,
			}, func() {
				m.setState(id, addrs, modified)
				m.getSem(addrs).Unlock()
			}, m.getSem(addrs)
	default:
		panic(state)
	}
}

func (m *msi) getSem(addrs []int32) *comp.Sem {
	alignedAddr := getAlignedMemoryAddress(addrs)
	sem, exists := m.pendings[alignedAddr]
	if !exists {
		sem = &comp.Sem{}
		m.pendings[alignedAddr] = sem
	}
	return sem
}

func (m *msi) getState(id int, addrs []int32) msiState {
	e := msiEntry{
		id:          id,
		alignedAddr: getAlignedMemoryAddress(addrs),
	}
	return m.states[e]
}

func (m *msi) setState(id int, addrs []int32, state msiState) {
	e := msiEntry{
		id:          id,
		alignedAddr: getAlignedMemoryAddress(addrs),
	}
	m.states[e] = state
	switch state {
	case invalid:
		//fmt.Println("transition", e.id, "invalid")
	case shared:
		//fmt.Println("transition", e.id, "shared")
	case modified:
		//fmt.Println("transition", e.id, "modified")
	}
}
