package types

import "time"

// EventKind represents the type of a parsed log event.
type EventKind int

const (
	EventUnknown EventKind = iota
	EventMapStart
	EventMapEnd
	EventBagInit
	EventBagMod
)

func (k EventKind) String() string {
	switch k {
	case EventMapStart:
		return "MapStart"
	case EventMapEnd:
		return "MapEnd"
	case EventBagInit:
		return "BagInit"
	case EventBagMod:
		return "BagMod"
	default:
		return "Unknown"
	}
}

// BagEvent captures inventory slot values from the log.
type BagEvent struct {
	PageID       int
	SlotID       int
	ConfigBaseID int
	Num          int
}

// Event is a normalized parsed log event.
type Event struct {
	Kind EventKind
	Time time.Time
	Line string // original line
	Bag  *BagEvent
}
