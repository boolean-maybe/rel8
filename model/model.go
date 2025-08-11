package model

import (
	"github.com/gdamore/tcell/v2"
	"rel8/db"
)

type State struct {
	Mode Mode
	// in table mode
	TableMode         TableMode
	TableHeaders      []string
	TableData         []db.TableData
	SelectedDataIndex int

	// in details mode
	DetailText string

	CommandText string
}

var Quit = &State{Mode: QuitMode} // Use special mode to identify quit state
var Initial = &State{Mode: Browse}

type Mode int

const (
	Browse Mode = iota
	Command
	SQL
	Detail
	QuitMode Mode = -1
)

type TableMode int

const (
	EmptyTable TableMode = iota
	DatabaseTable
	Database
	TableRow
)

type Event struct {
	Event *tcell.EventKey
	Text  string
	Row   int
}
