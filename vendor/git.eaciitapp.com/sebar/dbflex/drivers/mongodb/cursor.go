package mongodb

import (
	"reflect"

	"git.eaciitapp.com/sebar/dbflex"
	"github.com/eaciit/toolkit"
	mgo "gopkg.in/mgo.v2"
)

type Cursor struct {
	dbflex.CursorBase
	mgocursor *mgo.Query
	mgoiter   *mgo.Iter
	mgopipe   *mgo.Pipe

	isPipe bool
}

func (c *Cursor) Reset() error {
	if c.mgocursor == nil {
		return toolkit.Error("Cursor is not properly initialized")
	}
	if c.mgoiter != nil {
		c.mgoiter.Close()
	}

	if c.isPipe {
		c.mgoiter = c.mgopipe.Iter()
	} else {
		c.mgoiter = c.mgocursor.Iter()
	}
	return nil
}

func (c *Cursor) Fetch(result interface{}) error {
	if c.mgoiter == nil {
		return toolkit.Error("Cursor is not yet properly initialized")
	}
	ok := c.mgoiter.Next(result)
	if !ok {
		return toolkit.Error("EOF")
	}
	if c.CloseAfterFetch() {
		c.Close()
	}
	return nil
}

func (c *Cursor) Count() int {
	if c.mgocursor == nil {
		return 0
	}
	n, _ := c.mgocursor.Count()
	return n
}

func (c *Cursor) Fetchs(result interface{}, n int) error {
	defer func() {
		if c.CloseAfterFetch() {
			c.Close()
		}
	}()

	if c.mgoiter == nil {
		return toolkit.Error("Cursor is not yet properly initialized")
	}

	if n == 0 {
		err := c.mgoiter.All(result)
		return err
	} else {
		fetched := 0
		fetching := true

		v := reflect.TypeOf(result).Elem().Elem()
		ivs := reflect.MakeSlice(reflect.SliceOf(v), 0, 0)
		for fetching {
			iv := reflect.New(v).Interface()

			tiv := toolkit.M{}
			bOk := c.mgoiter.Next(&tiv)
			if bOk {
				toolkit.Serde(tiv, iv, "json")
				ivs = reflect.Append(ivs, reflect.ValueOf(iv).Elem())
				fetched++
				if fetched == n {
					fetching = false
				}
			} else {
				fetching = false
			}
		}
		reflect.ValueOf(result).Elem().Set(ivs)
	}
	return nil
}

func (c *Cursor) Close() {
	if c.mgoiter != nil {
		c.mgoiter.Close()
	}
}
