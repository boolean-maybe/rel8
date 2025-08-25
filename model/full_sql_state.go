package model

type FullSqlState struct {
	*CommonState
	Sql string
}
