package gora

import (
	"database/sql"

	_ "gopkg.in/goracle.v2"

	"git.eaciitapp.com/sebar/dbflex"
	"github.com/eaciit/toolkit"

	"git.eaciitapp.com/sebar/dbflex/drivers/rdbms"
)

// Connection implementation of dbflex.IConnection
type Connection struct {
	rdbms.Connection
	db *sql.DB
}

func init() {
	dbflex.RegisterDriver("oracle", func(si *dbflex.ServerInfo) dbflex.IConnection {
		c := new(Connection)
		c.SetThis(c)
		c.ServerInfo = *si
		return c
	})
}

// Connect to database instance
func (c *Connection) Connect() error {
	sqlconnstring := toolkit.Sprintf("%s/%s@%s/%s",
		c.User, c.Password,
		c.Host, c.Database)
	db, err := sql.Open("goracle", sqlconnstring)
	c.db = db
	return err
}

func (c *Connection) State() string {
	if c.db != nil {
		return dbflex.StateConnected
	}
	return dbflex.StateUnknown
}

// Close database connection
func (c *Connection) Close() {
	if c.db != nil {
		c.db.Close()
	}
}

// NewQuery generates new query object to perform query action
func (c *Connection) NewQuery() dbflex.IQuery {
	q := new(Query)
	q.SetThis(q)
	q.db = c.db
	return q
}
