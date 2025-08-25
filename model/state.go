package model

import (
	"github.com/gdamore/tcell/v2"
)

type Action interface {
}

type State interface {
	GetHeaderInfo() *HeaderInfo

	HasCommand() bool
	GetCommandState() CommandState

	HasBrowse() bool
	GetBrowseState() BrowseState

	HasSql() bool
	GetSqlState() SqlState

	HasTree() bool
	GetTreeState() TreeState

	HasDetail() bool
	GetDetailState() DetailState

	HasFullSql() bool
	GetFullSqlState() FullSqlState
}

type Event struct {
	EventType EventType
	Event     *tcell.EventKey
	Text      string
	Row       int
}

type EventType int

const (
	Init EventType = iota
	Quit
	Other
)

var InitEvent = &Event{
	EventType: Init,
}

var QuitEvent = &Event{
	EventType: Init,
}
