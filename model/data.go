package model

import "rel8/db"

type TableInfo struct {
	TableHeaders      []string
	TableData         []db.TableData
	SelectedDataIndex int
}

type HeaderInfo map[string]string
