package model

import "rel8/db"

type BrowseClass int

const (
	None BrowseClass = iota
	EmptyTable
	DatabaseTable
	View
	Procedure
	Function
	Trigger
	Database
	TableRow
)

type TableInfo struct {
	TableHeaders      []string
	TableData         []db.TableData
	SelectedDataIndex int
}

type BrowseState struct {
	*CommonState
	BrowseClass BrowseClass
	TableInfo   *TableInfo
}
