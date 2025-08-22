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

type HeaderInfo map[string]string

type BrowseState struct {
	BrowseClass BrowseClass
	TableInfo   *TableInfo
	HeaderInfo  *HeaderInfo
}
