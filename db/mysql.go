package db

import (
	"database/sql"
)

type TableData interface{}

type Mysql struct {
	DbInstance *sql.DB
}

type MysqlTable struct {
	Name   string
	Type   string
	Engine string
	Rows   string
	Size   string
}

type MysqlDatabase struct {
	Name      string
	Charset   string
	Collation string
}

func (m *Mysql) Db() *sql.DB {
	return m.DbInstance
}
