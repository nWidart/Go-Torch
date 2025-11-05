package tracker

import (
	"sync"
	"time"

	"GoTorch/internal/types"
)

type slotKey struct {
	PageID       int
	SlotID       int
	ConfigBaseID int
}

type MapSession struct {
	StartedAt time.Time
	EndedAt   time.Time
	Active    bool
	// Tally by ConfigBaseID -> total picked up during this session
	Tally map[int]int
}

type State struct {
	InMap      bool
	Current    MapSession
	TotalDrops int
	LastEvents []types.Event
	Inventory  map[slotKey]int // latest known counts per slot+item
}

type Tracker struct {
	mu    sync.Mutex
	state State
	// configuration knobs may go here later (filters, value tables)
}

func New() *Tracker {
	return &Tracker{state: State{Inventory: make(map[slotKey]int)}}
}

// GetState returns a snapshot copy of current state for use by UI/CLI.
func (t *Tracker) GetState() State {
	t.mu.Lock()
	defer t.mu.Unlock()
	// Deep-ish copy for safe reading
	st := State{
		InMap:      t.state.InMap,
		TotalDrops: t.state.TotalDrops,
		Inventory:  make(map[slotKey]int, len(t.state.Inventory)),
		LastEvents: make([]types.Event, len(t.state.LastEvents)),
		Current: MapSession{
			StartedAt: t.state.Current.StartedAt,
			EndedAt:   t.state.Current.EndedAt,
			Active:    t.state.Current.Active,
			Tally:     make(map[int]int, len(t.state.Current.Tally)),
		},
	}
	for k, v := range t.state.Inventory {
		st.Inventory[k] = v
	}
	copy(st.LastEvents, t.state.LastEvents)
	for k, v := range t.state.Current.Tally {
		st.Current.Tally[k] = v
	}
	return st
}

func (t *Tracker) appendEvent(ev types.Event) {
	const max = 100
	t.state.LastEvents = append(t.state.LastEvents, ev)
	if len(t.state.LastEvents) > max {
		// drop oldest
		copy(t.state.LastEvents, t.state.LastEvents[len(t.state.LastEvents)-max:])
		t.state.LastEvents = t.state.LastEvents[:max]
	}
}

// OnEvent ingests a parsed log event and updates state.
func (t *Tracker) OnEvent(ev *types.Event) {
	if ev == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	// record for debug view
	t.appendEvent(*ev)

	switch ev.Kind {
	case types.EventMapStart:
		// Start a new session; if one is active, finalize it first
		if t.state.InMap && t.state.Current.Active {
			s := t.state.Current
			s.Active = false
			s.EndedAt = ev.Time
			t.state.Current = s
		}
		t.state.InMap = true
		t.state.Current = MapSession{StartedAt: ev.Time, Active: true, Tally: make(map[int]int)}
	case types.EventMapEnd:
		if t.state.InMap {
			t.state.InMap = false
			s := t.state.Current
			s.Active = false
			s.EndedAt = ev.Time
			t.state.Current = s
		}
	case types.EventBagInit:
		if ev.Bag == nil {
			return
		}
		// Initialize/refresh inventory snapshot but do not count towards drops.
		key := slotKey{PageID: ev.Bag.PageID, SlotID: ev.Bag.SlotID, ConfigBaseID: ev.Bag.ConfigBaseID}
		t.state.Inventory[key] = ev.Bag.Num
	case types.EventBagMod:
		if ev.Bag == nil {
			return
		}
		key := slotKey{PageID: ev.Bag.PageID, SlotID: ev.Bag.SlotID, ConfigBaseID: ev.Bag.ConfigBaseID}
		prev := t.state.Inventory[key]
		delta := ev.Bag.Num - prev
		// Update inventory regardless of map state
		t.state.Inventory[key] = ev.Bag.Num
		// Count only positive increments while inside a map
		if delta > 0 && t.state.InMap && t.state.Current.Active {
			t.state.Current.Tally[ev.Bag.ConfigBaseID] += delta
			t.state.TotalDrops += delta
		}
	}
}
