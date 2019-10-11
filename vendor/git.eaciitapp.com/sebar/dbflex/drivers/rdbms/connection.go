package rdbms

import (
	"git.eaciitapp.com/sebar/dbflex"
)

type Connection struct {
	dbflex.ConnectionBase
}

func (c *Connection) Connect() error {
	panic("not implemented")
}

func (c *Connection) State() string {
	panic("not implemented")
}

func (c *Connection) Close() {
	panic("not implemented")
}

func (c *Connection) NewQuery() {
	panic("not implemented")
}

func (c *Connection) ObjectNames(dbflex.ObjTypeEnum) []string {
	panic("not implemented")
}
