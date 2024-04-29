package mvp7_0

import (
	"github.com/teivah/majorana/proc/comp"
)

type msiState = int32

const (
	invalid msiState = iota // Default has to be invalid
	shared
	modified
)

func (r *commandInfo) isDone() bool {
	return r.doneFlag
}

func (r *commandInfo) done() {
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
	pendings []*commandInfo
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

type commandRequest struct {
	id          int
	alignedAddr int32
	request     requestType
}

// TODO Ugly
type commandInfo struct {
	doneFlag    bool
	callback    func()
	alignedAddr int32
	request     requestType
}

type msi struct {
	pendings map[int32]*comp.Sem
	states   map[msiEntry]msiState
	commands map[commandRequest]*commandInfo
}

func newMSI() *msi {
	return &msi{
		pendings: make(map[int32]*comp.Sem),
		states:   make(map[msiEntry]msiState),
		commands: make(map[commandRequest]*commandInfo),
	}
}

var noop = func() {}

func (m *msi) requests(id int) []*commandInfo {
	var requests []*commandInfo
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
	var pendings []*commandInfo
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
				pendings = append(pendings, m.addCommand(e.id, alignedAddr, writeBack))
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

func (m *msi) addCommand(id int, alignedAddr int32, request requestType) *commandInfo {
	entry := msiEntry{
		id:          id,
		alignedAddr: alignedAddr,
	}
	commandEntry := commandRequest{
		id:          id,
		alignedAddr: alignedAddr,
		request:     request,
	}
	r := &commandInfo{
		callback: func() {
			m.states[entry] = invalid
			delete(m.commands, commandEntry)
		},
		alignedAddr: alignedAddr,
		request:     request,
	}
	if v, exists := m.commands[commandEntry]; exists {
		if v.alignedAddr != r.alignedAddr || v.request != r.request {
			panic("invalid state")
		}
		m.commands[commandEntry] = v
		return v
	} else {
		m.commands[commandEntry] = r
		return r
	}
}

// pendings: pre
// msiResponse: action
// callback: post
func (m *msi) lock(id int, addrs []int32) (msiResponse, func(), *comp.Sem) {
	alignedAddr := getAlignedMemoryAddress(addrs)
	state := m.getState(id, addrs)
	var pendings []*commandInfo
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
				pendings = append(pendings, m.addCommand(e.id, alignedAddr, writeBack))
			case shared:
				pendings = append(pendings, m.addCommand(e.id, alignedAddr, evict))
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
				pendings = append(pendings, m.addCommand(e.id, alignedAddr, writeBack))
			case shared:
				pendings = append(pendings, m.addCommand(e.id, alignedAddr, evict))
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

func (m *msi) evictExtraCacheLine(id int, alignedAddr int32) *commandInfo {
	state := m.states[msiEntry{
		id:          id,
		alignedAddr: alignedAddr,
	}]
	switch state {
	case shared:
		return m.addCommand(id, alignedAddr, evict)
	case modified:
		return m.addCommand(id, alignedAddr, writeBack)
	default:
		return nil
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
