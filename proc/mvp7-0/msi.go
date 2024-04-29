package mvp7_0

import (
	"github.com/teivah/majorana/proc/comp"
)

// msiState represents the different state that can be taken by a cache line per
// core.
type msiState = int32

const (
	invalid msiState = iota // Default has to be invalid
	shared
	modified
)

type requestType = int32

const (
	// Make sure a zero value isn't confused with an element
	evict requestType = iota + 1
	writeBack
)

// msiResponse represents the response to an MSI query.
type msiResponse struct {
	// pendings are the actions that have to be completed before the query
	pendings []*msiCommandInfo

	// Mutually exclusive
	// wait means can't lock for now
	wait bool
	// fetchFromMemory means fetch from memory then store into L1
	fetchFromMemory bool
	// readFromL1 means the line is already fetched, the core can read from L1
	readFromL1 bool
	// writeToL1 means the line is already fetched, the core can write to L1
	writeToL1 bool
}

type msi struct {
	pendings map[int32]*comp.Sem
	states   map[msiEntry]msiState
	commands map[msiCommandRequest]*msiCommandInfo
}

type msiEntry struct {
	id          int
	alignedAddr int32
}

// msiCommandRequest is a request to a specific core (snoop)
type msiCommandRequest struct {
	id          int
	alignedAddr int32
	request     requestType
}

// msiCommandInfo represents an additional source of information to a msiCommandRequest
type msiCommandInfo struct {
	doneFlag bool
	callback func()
	request  requestType
}

// isDone tells whether the command is completed
func (r *msiCommandInfo) isDone() bool {
	return r.doneFlag
}

// done completes a command
func (r *msiCommandInfo) done() {
	r.doneFlag = true
	r.callback()
}

func newMSI() *msi {
	return &msi{
		pendings: make(map[int32]*comp.Sem),
		states:   make(map[msiEntry]msiState),
		commands: make(map[msiCommandRequest]*msiCommandInfo),
	}
}

var noop = func() {}

// getPendingRequestsToCore gets the pending requests to a specific core (snoop)
func (m *msi) getPendingRequestsToCore(id int) map[msiCommandRequest]*msiCommandInfo {
	requests := make(map[msiCommandRequest]*msiCommandInfo)
	for req, info := range m.commands {
		if req.id != id {
			continue
		}
		requests[req] = info
	}
	return requests
}

// rLock is a lock for read
// Workflows:
// Pre-actions: pendings
// Action: msiResponse
// Post-action: msiCommandInfo callback
func (m *msi) rLock(id int, addrs []int32) (msiResponse, func(), *comp.Sem) {
	alignedAddr := getAlignedMemoryAddress(addrs)
	state := m.getState(id, addrs)
	switch state {
	case invalid:
		if !m.getSem(addrs).RLock() {
			return msiResponse{wait: true}, noop, nil
		}
		pendings := m.readRequest(id, alignedAddr)
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

// readRequest means a core with an invalid line wants to read from it
func (m *msi) readRequest(id int, alignedAddr int32) []*msiCommandInfo {
	var pendings []*msiCommandInfo
	for e, state := range m.states {
		if id == e.id {
			continue
		}
		if alignedAddr != e.alignedAddr {
			continue
		}
		switch state {
		case modified:
			pendings = append(pendings, m.sendNewMSICommand(e.id, alignedAddr, writeBack))
		}
	}
	return pendings
}

// lock is a lock for write
// Workflows:
// Pre-actions: pendings
// Action: msiResponse
// Post-action: msiCommandInfo callback
func (m *msi) lock(id int, addrs []int32) (msiResponse, func(), *comp.Sem) {
	alignedAddr := getAlignedMemoryAddress(addrs)
	state := m.getState(id, addrs)
	switch state {
	case invalid:
		if !m.getSem(addrs).Lock() {
			return msiResponse{wait: true}, noop, nil
		}

		pendings := m.writeRequest(id, alignedAddr)
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

		pendings := m.invalidationRequest(id, alignedAddr)
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

// writeRequest means a core with an invalid line wants to write to it
func (m *msi) writeRequest(id int, alignedAddr int32) []*msiCommandInfo {
	var pendings []*msiCommandInfo
	for e, state := range m.states {
		if id == e.id {
			continue
		}
		if alignedAddr != e.alignedAddr {
			continue
		}
		switch state {
		case modified:
			pendings = append(pendings, m.sendNewMSICommand(e.id, alignedAddr, writeBack))
		case shared:
			pendings = append(pendings, m.sendNewMSICommand(e.id, alignedAddr, evict))
		}
	}
	return pendings
}

// invalidationRequest means a core with an shared line wants to write to it
func (m *msi) invalidationRequest(id int, alignedAddr int32) []*msiCommandInfo {
	var pendings []*msiCommandInfo
	for e, state := range m.states {
		if id == e.id {
			continue
		}
		if alignedAddr != e.alignedAddr {
			continue
		}
		// We want to evict the line with or without write-back first
		switch state {
		case modified:
			pendings = append(pendings, m.sendNewMSICommand(e.id, alignedAddr, writeBack))
		case shared:
			pendings = append(pendings, m.sendNewMSICommand(e.id, alignedAddr, evict))
		}
	}
	return pendings
}

// evictExtraCacheLine evicts a cache line when L1 is full
func (m *msi) evictExtraCacheLine(id int, alignedAddr int32) *msiCommandInfo {
	state := m.states[msiEntry{
		id:          id,
		alignedAddr: alignedAddr,
	}]
	switch state {
	case shared:
		return m.sendNewMSICommand(id, alignedAddr, evict)
	case modified:
		return m.sendNewMSICommand(id, alignedAddr, writeBack)
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
}

// sendNewMSICommand sends a new MSI command to a specific core (snoop)
func (m *msi) sendNewMSICommand(id int, alignedAddr int32, request requestType) *msiCommandInfo {
	cmdRequest := msiCommandRequest{
		id:          id,
		alignedAddr: alignedAddr,
		request:     request,
	}
	if existingCommand, exists := m.commands[cmdRequest]; exists {
		if existingCommand.request != request {
			panic("invalid state")
		}
		// It means a similar command was already issued and not yet completed
		// In this case, we don't create a new command, we reuse the pending one
		m.commands[cmdRequest] = existingCommand
		return existingCommand
	} else {
		newCommand := &msiCommandInfo{
			callback: func() {
				m.states[msiEntry{id, alignedAddr}] = invalid
				delete(m.commands, cmdRequest)
			},
			request: request,
		}
		m.commands[cmdRequest] = newCommand
		return newCommand
	}
}
