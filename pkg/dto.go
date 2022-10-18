package game

import "fmt"

type EventType int

const (
	NoOp EventType = iota
	CursorMove
	OpenCell
)

type Event struct {
	Type     EventType
	Position Point
}

func NewEvent(t EventType, pos Point) *Event {
	return &Event{
		Type:    t,
		Position: pos,
	}
}

func NewEventFromBytes(bs []byte) *Event {
	e := new(Event)
	FromGob(bs, e)
	return e
}

func (e *Event) String() string {
	titles := []string{
		"NoOp",
		"CursorMove",
		"OpenCell",
	}
	return fmt.Sprintf("[%s] %v", titles[e.Type], e.Position)
}

func (e *Event) Bytes() []byte {
	return ToGob(*e)
}
