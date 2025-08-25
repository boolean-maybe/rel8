package model

type SqlState struct {
	*CommonState
	Sql string
}
