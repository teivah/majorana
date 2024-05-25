package mvp8_0

import (
	"fmt"
	"sync"

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
	l1Evict requestType = iota + 1
	l1WriteBack
	l3Evict
	l3WriteBack
)

// msiResponse represents the response to an MSI query.
type msiResponse struct {
	// pendings are the actions that have to be completed before the query
	pendings []*msiCommandInfo

	// Mutually exclusive
	// wait means can't l1Lock for now
	wait bool
	// notFromL1 means fetch from memory then store into L1
	notFromL1 bool
	// fromL1 means the line is already fetched, the core can read from L1
	fromL1 bool
	// writeToL1 means the line is already fetched, the core can write to L1
	writeToL1 bool
}

type msi struct {
	pendings map[comp.AlignedAddress]*comp.Sem
	states   map[msiEntry]msiState
	// An eviction happened, the CU has to synchronize the state
	staleState bool
	commands   map[msiCommandRequest]*msiCommandInfo
	// A line is locked when it's being fetched
	l3Lock map[comp.AlignedAddress]*sync.Mutex
	// Indicates whether an L3 line is pending write (used to know whether a
	// cache eviction should be a simple eviction or a write-back)
	l3Write map[comp.AlignedAddress]bool

	// Monitoring
	l1EvictRequestCount     int
	l1WriteBackRequestCount int
	l3EvictRequestCount     int
	l3WriteBackRequestCount int
}

type msiEntry struct {
	id          int
	alignedAddr comp.AlignedAddress
}

func (msiEntry) less() []func(msiEntry) int {
	return []func(msiEntry) int{
		func(m msiEntry) int { return m.id },
		func(m msiEntry) int { return int(m.alignedAddr) },
	}
}

// msiCommandRequest is a request to a specific core (snoop)
type msiCommandRequest struct {
	id          int
	alignedAddr comp.AlignedAddress
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
		pendings: make(map[comp.AlignedAddress]*comp.Sem),
		states:   make(map[msiEntry]msiState),
		commands: make(map[msiCommandRequest]*msiCommandInfo),
		l3Lock:   make(map[comp.AlignedAddress]*sync.Mutex),
		l3Write:  make(map[comp.AlignedAddress]bool),
	}
}

func (m *msi) copyState() map[msiEntry]msiState {
	res := make(map[msiEntry]msiState, len(m.states))
	for k, v := range m.states {
		res[k] = v
	}
	return res
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

// l1RLock is a lock for read
// Workflows:
// Pre-actions: pendings
// Action: msiResponse
// Post-action: msiCommandInfo callback
func (m *msi) l1RLock(id int, addrs []int32) (msiResponse, func(), *comp.Sem) {
	alignedAddr := getL1AlignedMemoryAddress(addrs)
	state := m.getL1State(id, addrs)
	switch state {
	case invalid:
		if !m.getL1Sem(addrs).RLock() {
			return msiResponse{wait: true}, noop, nil
		}
		pendings := m.l1ReadRequest(id, alignedAddr)
		return msiResponse{
				notFromL1: true,
				pendings:  pendings,
			}, func() {
				m.setL1State(id, addrs, shared)
				m.getL1Sem(addrs).RUnlock()
			}, m.getL1Sem(addrs)
	case modified:
		if !m.getL1Sem(addrs).Lock() {
			return msiResponse{wait: true}, noop, nil
		}
		return msiResponse{fromL1: true}, func() {
			m.getL1Sem(addrs).Unlock()
		}, m.getL1Sem(addrs)
	case shared:
		if !m.getL1Sem(addrs).RLock() {
			return msiResponse{wait: true}, noop, nil
		}
		return msiResponse{fromL1: true}, func() {
			m.getL1Sem(addrs).RUnlock()
		}, m.getL1Sem(addrs)
	default:
		panic(state)
	}
}

// l1ReadRequest means a core with an invalid line wants to read from it
func (m *msi) l1ReadRequest(id int, alignedAddr comp.AlignedAddress) []*msiCommandInfo {
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
			pendings = append(pendings, m.sendNewL1MSICommand(e.id, alignedAddr, l1WriteBack))
		}
	}
	return pendings
}

// l1Lock is a lock for write
// Workflows:
// Pre-actions: pendings
// Action: msiResponse
// Post-action: msiCommandInfo callback
func (m *msi) l1Lock(id int, addrs []int32) (msiResponse, func(), *comp.Sem) {
	alignedAddr := getL1AlignedMemoryAddress(addrs)
	state := m.getL1State(id, addrs)
	switch state {
	case invalid:
		if !m.getL1Sem(addrs).Lock() {
			return msiResponse{wait: true}, noop, nil
		}

		pendings := m.l1WriteRequest(id, alignedAddr)
		return msiResponse{
				notFromL1: true,
				pendings:  pendings,
			}, func() {
				m.setL1State(id, addrs, modified)
				m.getL1Sem(addrs).Unlock()
			}, m.getL1Sem(addrs)
	case modified:
		if !m.getL1Sem(addrs).Lock() {
			return msiResponse{wait: true}, noop, nil
		}
		return msiResponse{writeToL1: true}, func() {
			m.getL1Sem(addrs).Unlock()
		}, m.getL1Sem(addrs)
	case shared:
		if !m.getL1Sem(addrs).Lock() {
			return msiResponse{wait: true}, noop, nil
		}

		pendings := m.l1InvalidationRequest(id, alignedAddr)
		return msiResponse{
				writeToL1: true,
				pendings:  pendings,
			}, func() {
				m.setL1State(id, addrs, modified)
				m.getL1Sem(addrs).Unlock()
			}, m.getL1Sem(addrs)
	default:
		panic(state)
	}
}

// l1WriteRequest means a core with an invalid line wants to write to it
func (m *msi) l1WriteRequest(id int, alignedAddr comp.AlignedAddress) []*msiCommandInfo {
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
			pendings = append(pendings, m.sendNewL1MSICommand(e.id, alignedAddr, l1WriteBack))
		case shared:
			pendings = append(pendings, m.sendNewL1MSICommand(e.id, alignedAddr, l1Evict))
		}
	}
	return pendings
}

// l1InvalidationRequest means a core with a shared line wants to write to it
func (m *msi) l1InvalidationRequest(id int, alignedAddr comp.AlignedAddress) []*msiCommandInfo {
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
			pendings = append(pendings, m.sendNewL1MSICommand(e.id, alignedAddr, l1WriteBack))
		case shared:
			pendings = append(pendings, m.sendNewL1MSICommand(e.id, alignedAddr, l1Evict))
		}
	}
	return pendings
}

// evictL1ExtraCacheLine evicts a cache line when L1 is full
func (m *msi) evictL1ExtraCacheLine(id int, alignedAddr comp.AlignedAddress) *msiCommandInfo {
	state := m.states[msiEntry{
		id:          id,
		alignedAddr: alignedAddr,
	}]
	switch state {
	case shared, invalid:
		return m.sendNewL1MSICommand(id, alignedAddr, l1Evict)
	case modified:
		return m.sendNewL1MSICommand(id, alignedAddr, l1WriteBack)
	default:
		panic(fmt.Sprintf("unknown %d", state))
	}
}

func (m *msi) evictL3ExtraCacheLine(id int, alignedAddr comp.AlignedAddress) *msiCommandInfo {
	if m.l3Write[alignedAddr] {
		m.l3WriteBackRequestCount++
		return m.sendNewL3MSICommand(id, alignedAddr, l3WriteBack)
	}
	m.l3EvictRequestCount++
	return m.sendNewL3MSICommand(id, alignedAddr, l3Evict)
}

func (m *msi) getL1Sem(addrs []int32) *comp.Sem {
	alignedAddr := getL1AlignedMemoryAddress(addrs)
	sem, exists := m.pendings[alignedAddr]
	if !exists {
		sem = &comp.Sem{}
		m.pendings[alignedAddr] = sem
	}
	return sem
}

func (m *msi) getL1State(id int, addrs []int32) msiState {
	e := msiEntry{
		id:          id,
		alignedAddr: getL1AlignedMemoryAddress(addrs),
	}
	return m.states[e]
}

func (m *msi) setL1State(id int, addrs []int32, state msiState) {
	e := msiEntry{
		id:          id,
		alignedAddr: getL1AlignedMemoryAddress(addrs),
	}
	m.states[e] = state
	m.staleState = true
}

// sendNewL1MSICommand sends a new MSI command to a specific core (snoop)
func (m *msi) sendNewL1MSICommand(id int, alignedAddr comp.AlignedAddress, request requestType) *msiCommandInfo {
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
		if request == l1Evict {
			m.l1EvictRequestCount++
		} else if request == l1WriteBack {
			m.l1WriteBackRequestCount++
		}
		return newCommand
	}
}

// sendNewL3MSICommand sends a new MSI command to a specific core (snoop)
func (m *msi) sendNewL3MSICommand(id int, alignedAddr comp.AlignedAddress, request requestType) *msiCommandInfo {
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
				delete(m.commands, cmdRequest)
			},
			request: request,
		}
		m.commands[cmdRequest] = newCommand
		return newCommand
	}
}

func (m *msi) getL3Lock(addrs []int32) *sync.Mutex {
	addr := getL3AlignedMemoryAddress(addrs)
	mu, exists := m.l3Lock[addr]
	if !exists {
		mu = &sync.Mutex{}
		m.l3Lock[addr] = mu
	}
	return mu
}

func (m *msi) l3WriteNotify(addr comp.AlignedAddress) {
	m.l3Write[addr] = true
}

func (m *msi) l3ReleaseWriteNotify(addr comp.AlignedAddress) {
	m.l3Write[addr] = false
}

func (m *msi) stats() map[string]any {
	return map[string]any{
		"msi_l1_evict_request":     m.l1EvictRequestCount,
		"msi_l1_writeback_request": m.l1WriteBackRequestCount,
		"msi_l3_evict_request":     m.l3EvictRequestCount,
		"msi_l3_writeback_request": m.l3WriteBackRequestCount,
	}
}
