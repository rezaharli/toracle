package gora

import (
	"git.eaciitapp.com/sebar/dbflex/drivers/rdbms"
	"github.com/eaciit/toolkit"
)

// Cursor represent cursor object. Inherits Cursor object of rdbms drivers and implementation of dbflex.ICursor
type Cursor struct {
	rdbms.Cursor
	fieldNames []string
	fieldTypes []string
}

func (c *Cursor) Serialize(
	dataList, fieldMap toolkit.M, fieldValues []interface{}, dest interface{}) error {
	var err error
	mobj := toolkit.M{}
	toolkit.Serde(dest, &mobj, "")

	values := c.Values()
	for idx, fieldName := range c.fieldNames {
		//fmt.Println(idx, ":", fieldName, ":", c.fieldTypes[idx], "=", values[idx])
		switch c.fieldTypes[idx] {
		case "int":
			mobj.Set(fieldName, toolkit.ToInt(values[idx], toolkit.RoundingAuto))

		case "float64":
			f64 := toolkit.ToFloat64(values[idx], 4, toolkit.RoundingAuto)
			mobj.Set(fieldName, f64)

		case "time.Time":
			mobj.Set(fieldName, values[idx])

		default:
			mobj.Set(fieldName, values[idx])
		}
	}

	err = toolkit.Serde(mobj, dest, "")
	if err != nil {
		return toolkit.Error(err.Error() + toolkit.Sprintf(" object: %s", toolkit.JsonString(mobj)))
	}
	return nil
}
