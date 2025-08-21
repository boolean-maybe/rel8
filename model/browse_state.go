package model

import "rel8/db"

type BrowseState struct {
	class             Class
	tableHeaders      []string
	tableData         []db.TableData
	selectedDataIndex int
}

func (b *BrowseState) GetMode() Mode {
	return Mode{Kind: Browse, Class: b.class}
}

func (b *BrowseState) GetData() interface{} {
	return struct {
		TableHeaders      []string
		TableData         []db.TableData
		SelectedDataIndex int
	}{b.tableHeaders, b.tableData, b.selectedDataIndex}
}

func (b *BrowseState) GetAction() []Action {
	return []Action{}
}
