package rdbms

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/eaciit/toolkit"

	"git.eaciitapp.com/sebar/dbflex"
)

type RdbmsCursor interface {
	SerializeFieldType(string, reflect.Type, interface{}) (interface{}, error)
	Serialize(toolkit.M, toolkit.M, []interface{}, interface{}) error
	SerializeField(string, interface{}) (interface{}, error)
	Values() []interface{}
	ValuesPtr() []interface{}
	//GetValueType([]byte) reflect.Type
}

type Cursor struct {
	dbflex.CursorBase
	fetcher   *sql.Rows
	dest      []interface{}
	columns   []string
	values    []interface{}
	valuesPtr []interface{}
	m         toolkit.M

	_this dbflex.ICursor

	query        RdbmsQuery
	dataTypeList toolkit.M
}

func (c *Cursor) Reset() error {
	c.fetcher = nil
	c.dest = []interface{}{}
	return nil
}

func (c *Cursor) SetFetcher(r *sql.Rows) error {
	c.fetcher = r
	c.m = toolkit.M{}

	var err error
	c.columns, err = c.fetcher.Columns()
	if err != nil {
		return fmt.Errorf("unable to fetch columns. %s", err.Error())
	}

	count := len(c.columns)
	c.values = make([]interface{}, count)
	c.valuesPtr = make([]interface{}, count)

	for i, v := range c.columns {
		c.valuesPtr[i] = &c.values[i]
		c.m.Set(v, i)
	}
	return nil
}

func (c *Cursor) SetThis(ic dbflex.ICursor) dbflex.ICursor {
	c._this = ic
	return c
}

func (c *Cursor) this() dbflex.ICursor {
	return c._this
}

func (c *Cursor) Scan() error {
	if c.Error() != nil {
		return c.Error()
	}

	if c.fetcher == nil {
		return toolkit.Error("cursor is not valid, no fetcher object specified")
	}

	if !c.fetcher.Next() {
		return toolkit.Error("EOF")
	}

	return c.fetcher.Scan(c.valuesPtr...)
}

func (c *Cursor) Values() []interface{} {
	return c.values
}

func (c *Cursor) ValuesPtr() []interface{} {
	return c.valuesPtr
}

func (c *Cursor) SerializeFieldType(name string, dtype reflect.Type, value interface{}) (interface{}, error) {
	return nil, toolkit.Error("SerializeFieldType is not yet properlhy implemented")
}

func (c *Cursor) SerializeField(name string, value interface{}) (interface{}, error) {
	for k, dtype := range c.dataTypeList {
		if strings.ToLower(k) == strings.ToLower(name) {
			return c.this().(RdbmsCursor).SerializeFieldType(name, dtype.(reflect.Type), value)
		}
	}

	return nil, toolkit.Errorf("field or attribute %s could not be found", name)
}

func (c *Cursor) Serialize(
	dataList, fieldMap toolkit.M, fieldValues []interface{}, dest interface{}) error {
	var err error
	mobj := toolkit.M{}
	toolkit.Serde(dest, &mobj, "")

	//-- if dateTypeList if not yet created, create new one
	if len(dataList) == 0 {
		for k, v := range mobj {
			typeName := reflect.TypeOf(c.valuesPtr[toolkit.ToInt(v, toolkit.RoundingAuto)])
			//toolkit.Printfn("Type: %s", typeName)
			c.dataTypeList.Set(k, typeName)
		}
	}

	for k, v := range fieldMap {
		var vtr interface{}
		if vtr, err = c.this().(RdbmsCursor).SerializeField(k,
			string(fieldValues[toolkit.ToInt(v, toolkit.RoundingAuto)].([]byte))); err != nil {
			return err
		} else {
			mobj.Set(k, vtr)
		}
	}

	err = toolkit.Serde(mobj, dest, "")
	if err != nil {
		return toolkit.Error(err.Error() + toolkit.Sprintf(" object: %s", toolkit.JsonString(mobj)))
	}
	return nil
}

func (c *Cursor) getDataTypeString(name string) string {
	if c.dataTypeList == nil {
		c.dataTypeList = toolkit.M{}
	}
	t := c.dataTypeList.Get(name, nil)
	if t == nil {
		return ""
	} else {
		return t.(reflect.Type).String()
	}
}

func (c *Cursor) Fetch(obj interface{}) error {
	err := c.Scan()
	if err != nil {
		return err
	}
	c.getTypeList(obj)
	//toolkit.Printfn("TypeList: %s", toolkit.JsonString(c.dataTypeList))
	err = c.this().(RdbmsCursor).Serialize(c.dataTypeList, c.m, c.values, obj)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cursor) Fetchs(obj interface{}, n int) error {
	var err error

	//--- get first model
	if c.dataTypeList == nil || len(c.dataTypeList) == 0 {
		var single interface{}
		if single, err = toolkit.GetEmptySliceElement(obj); err != nil {
			return toolkit.Errorf("unable to get slice element: %s", err.Error())
		}
		if single == nil {
			return fmt.Errorf("output type is nil")
		}
		c.getTypeList(single)
	}

	i := 0
	loop := true
	ms := []toolkit.M{}
	for loop {
		err = c.Scan()
		if err != nil {
			if n == 0 && err.Error() == "EOF" {
				loop = false
				err = nil
			} else {
				return err
			}
		} else {
			mobj := toolkit.M{}
			err = c.this().(RdbmsCursor).Serialize(c.dataTypeList, c.m, c.values, &mobj)
			if err != nil {
				return err
			}
			ms = append(ms, mobj)
			i++
			if i == n {
				loop = false
			}
		}
	}

	err = toolkit.Serde(ms, obj, "")
	if err != nil {
		return err
	}
	return nil
}

func (c *Cursor) Close() {
	if c.fetcher != nil {
		c.fetcher.Close()
	}
}

func (c *Cursor) getTypeList(obj interface{}) {
	c.dataTypeList = toolkit.M{}
	fieldnames, fieldtypes, _, _ := ParseSQLMetadata(nil, obj)
	for i, name := range fieldnames {
		c.dataTypeList.Set(name, fieldtypes[i])
	}

	//-- obj is not accepted, retrieve via fetcher
	if len(fieldnames) == 0 {
		for k, v := range c.m {
			vindex := v.(int)
			if c.values[vindex] != nil {
				vbytes := c.values[vindex].([]byte)
				vtype := c.getValueType(vbytes)
				c.dataTypeList.Set(k, vtype)
			}
		}
	}
}

func (c *Cursor) getValueType(bs []byte) reflect.Type {
	var vtype reflect.Type
	value := string(bs)

	if _, e := toolkit.IsStringNumber(value, ""); e == nil {
		vtype = reflect.TypeOf(float64(0))
	} else {
		vtype = reflect.TypeOf("")
	}
	return vtype
}
