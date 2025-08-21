package model

import (
	"github.com/gdamore/tcell/v2"
)

type Action interface {
}

type Mode struct {
	Kind  Kind
	Class Class
}

type State interface {
	// GetMode state mode - part that defines the state type
	GetMode() Mode
	// GetData table data, headers, selected row index, details text etc.
	GetData() interface{}
	GetAction() []Action
}

type Kind int

const (
	Empty Kind = iota
	Tree
	Browse
	Command
	SQL
	Detail
	Editor
	QuitKind Kind = -1
)

type Class int

const (
	None Class = iota
	EmptyTable
	DatabaseTable
	View
	Procedure
	Function
	Trigger
	Database
	TableRow
)

var Initial = &StateAdapter{kind: Empty}
var Quit = &StateAdapter{kind: QuitKind}

type Event struct {
	Event *tcell.EventKey
	Text  string
	Row   int
}

type StateAdapter struct {
	kind Kind
}

func (s *StateAdapter) GetMode() Mode {
	return Mode{Kind: s.kind, Class: None}
}

func (s *StateAdapter) GetData() interface{} {
	return nil
}

func (s *StateAdapter) GetAction() []Action {
	return []Action{}
}
