package mongodb

import (
	"regexp"
	"time"

	"git.eaciitapp.com/sebar/dbflex"
	"github.com/eaciit/toolkit"
	mgo "gopkg.in/mgo.v2"
)

type Connection struct {
	dbflex.ConnectionBase

	mgosession *mgo.Session
}

func init() {
	dbflex.RegisterDriver("mongodb", func(si *dbflex.ServerInfo) dbflex.IConnection {
		c := new(Connection)
		c.ServerInfo = *si
		c.SetThis(c)
		c.SetFieldNameTag("bson")
		return c
	})
}

func (c *Connection) Connect() error {
	info := new(mgo.DialInfo)
	if c.User != "" {
		info.Username = c.User
		info.Password = c.Password
		info.Source = c.Config.GetString("authdb")
		if info.Source == "" {
			info.Source = "admin"
		}
	}
	info.Database = c.Database
	host := c.Host
	info.Addrs = []string{host}
	timeout := c.Config.GetInt("timeout")
	if timeout > 0 {
		info.Timeout = time.Duration(timeout) * time.Second
	}

	sess, e := mgo.DialWithInfo(info)
	if e != nil {
		return toolkit.Errorf("unable to connect: %s", e.Error())
	}
	c.mgosession = sess
	return nil
}

func (c *Connection) State() string {
	if c.mgosession != nil {
		return dbflex.StateConnected
	}
	return dbflex.StateUnknown
}

func (c *Connection) NewQuery() dbflex.IQuery {
	q := new(Query)
	q.SetThis(q)
	//q.session = c.mgosession
	q.db = c.mgosession.DB(c.Database)
	return q
}

func (c *Connection) ObjectNames(obj dbflex.ObjTypeEnum) []string {
	if c.mgosession == nil {
		return []string{}
	}

	mgoDb := c.mgosession.DB(c.Database)
	if obj == "" {
		obj = dbflex.ObjTypeAll
	}

	astr := []string{}

	if obj == dbflex.ObjTypeAll || obj == dbflex.ObjTypeTable {
		cols, err := mgoDb.CollectionNames()
		if err != nil {
			return []string{}
		}

		for _, col := range cols {
			if cond, _ := regexp.MatchString("^(.*)((\\.(indexes)|\\.(js)))$", col); !cond {
				astr = append(astr, col)
			}
		}

	}

	if obj == dbflex.ObjTypeAll || obj == dbflex.ObjTypeProcedure {
		cols := mgoDb.C("system.js")
		res := []toolkit.M{}
		err := cols.Find(nil).All(&res)
		if err != nil {
			toolkit.Printf("%v\n", err.Error())
			return []string{}
		}

		// toolkit.Printf("%v\n", res)
		for _, col := range res {
			astr = append(astr, col["_id"].(string))
		}

	}

	return astr
}

func (c *Connection) Close() {
	if c.mgosession != nil {
		c.mgosession.Close()
	}
}
