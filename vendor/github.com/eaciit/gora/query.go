package gora

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"git.eaciitapp.com/sebar/dbflex"
	"git.eaciitapp.com/sebar/dbflex/drivers/rdbms"
	"github.com/eaciit/toolkit"
)

// Query implementaion of dbflex.IQuery
type Query struct {
	rdbms.Query
	db         *sql.DB
	sqlcommand string
}

// Cursor produces a cursor from query
func (q *Query) Cursor(in toolkit.M) dbflex.ICursor {
	cursor := new(Cursor)
	cursor.SetThis(cursor)

	ct := q.Config(dbflex.ConfigKeyCommandType, dbflex.QuerySelect).(string)
	if ct != dbflex.QuerySelect && ct != dbflex.QuerySQL {
		cursor.SetError(toolkit.Errorf("cursor is used for only select command"))
		return cursor
	}

	cmdtxt := q.Config(dbflex.ConfigKeyCommand, "").(string)
	if cmdtxt == "" {
		cursor.SetError(toolkit.Errorf("no command"))
		return cursor
	}

	tablename := q.Config(dbflex.ConfigKeyTableName, "").(string)
	cq := dbflex.From(tablename).Select("count(*) as Count")
	if filter := q.Config(dbflex.ConfigKeyFilter, nil); filter != nil {
		cq.Where(filter.(*dbflex.Filter))
	}
	cursor.SetCountCommand(cq)

	//fmt.Println("query: " + cmdtxt)
	rows, err := q.db.Query(cmdtxt)
	if rows == nil {
		cursor.SetError(toolkit.Errorf("%s. SQL Command: %s", err.Error(), cmdtxt))
	} else {
		cursor.SetFetcher(rows)

		fieldnames, err := rows.Columns()
		if err != nil {
			cursor.SetError(toolkit.Errorf("unable to retrieve column names. %s. SQL: %s", err.Error(), cmdtxt))
			return cursor
		}
		cursor.fieldNames = fieldnames

		fieldtypes := []string{}
		cts, err := rows.ColumnTypes()
		if err != nil {
			cursor.SetError(toolkit.Errorf("unable to retrieve column type. %s. SQL: %s", err.Error(), cmdtxt))
			return cursor
		}

		for _, ct := range cts {
			if _, scale, ok := ct.DecimalSize(); ok && scale == 0 {
				fieldtypes = append(fieldtypes, "int")
			} else if ct.DatabaseTypeName() == "NUMBER" {
				fieldtypes = append(fieldtypes, "float64")
			} else if ct.DatabaseTypeName() == "DATE" {
				fieldtypes = append(fieldtypes, "time.Time")
			} else {
				fieldtypes = append(fieldtypes, "string")
			}
		}
		cursor.fieldTypes = fieldtypes
	}
	return cursor
}

// Execute will executes non-select command of a query
func (q *Query) Execute(in toolkit.M) (interface{}, error) {
	cmdtype, ok := q.Config(dbflex.ConfigKeyCommandType, dbflex.QuerySelect).(string)
	if !ok {
		return nil, toolkit.Errorf("Operation is unknown. current operation is %s", cmdtype)
	}
	cmdtxt := q.Config(dbflex.ConfigKeyCommand, "").(string)
	if cmdtxt == "" {
		return nil, toolkit.Errorf("No command")
	}

	var (
		sqlfieldnames []string
		sqlvalues     []string
	)

	data, hasData := in["data"]
	if !hasData && !(cmdtype == dbflex.QueryDelete || cmdtype == dbflex.QuerySelect) {
		return nil, toolkit.Error("non select and delete command should has data")
	}

	if hasData {
		sqlfieldnames, _, _, sqlvalues = rdbms.ParseSQLMetadata(q, data)
		affectedfields := q.Config("fields", []string{}).([]string)
		if len(affectedfields) > 0 {
			newfieldnames := []string{}
			newvalues := []string{}
			for idx, field := range sqlfieldnames {
				for _, find := range affectedfields {
					if strings.ToLower(field) == strings.ToLower(find) {
						newfieldnames = append(newfieldnames, find)
						newvalues = append(newvalues, sqlvalues[idx])
					}
				}
			}
			sqlfieldnames = newfieldnames
			sqlvalues = newvalues
		}
	}

	switch cmdtype {
	case dbflex.QueryInsert:
		cmdtxt = strings.Replace(cmdtxt, "{{.FIELDS}}", strings.Join(sqlfieldnames, ","), -1)
		cmdtxt = strings.Replace(cmdtxt, "{{.VALUES}}", strings.Join(sqlvalues, ","), -1)
		//toolkit.Printfn("\nCmd: %s", cmdtxt)

	case dbflex.QueryUpdate:
		//fmt.Println("fieldnames:", sqlfieldnames)
		updatedfields := []string{}
		for idx, fieldname := range sqlfieldnames {
			updatedfields = append(updatedfields, fieldname+"="+sqlvalues[idx])
		}
		cmdtxt = strings.Replace(cmdtxt, "{{.FIELDVALUES}}", strings.Join(updatedfields, ","), -1)
	}

	//fmt.Println("Cmd: ", cmdtxt)
	r, err := q.db.Exec(cmdtxt)

	if err != nil {
		return nil, toolkit.Errorf("%s. SQL Command: %s", err.Error(), cmdtxt)
	}
	return r, nil
}

// ExecType to identify type of exec
type ExecType int

const (
	ExecQuery ExecType = iota
	ExecNonQuery
	ExecQueryRow
)

/*
func (q *Query) SQL(string cmd, exec) dbflex.IQuery {
	swicth()
}
*/

func (q *Query) Templates() map[string]string {
	return map[string]string{
		string(dbflex.QuerySelect): "SELECT tmp.* FROM (SELECT {{.FIELDS}} FROM {{." + dbflex.ConfigKeyTableName + "}} " +
			"{{." + dbflex.QueryWhere + "}} " +
			"{{." + dbflex.QueryOrder + "}} " +
			"{{." + dbflex.QueryGroup + "}}) tmp " +
			"{{." + dbflex.QueryTake + "}}",
		//dbflex.QueryWhere: "{{." + dbflex.QueryWhere + "}}",
		dbflex.QueryTake:  "WHERE ROWNUM <= {{." + dbflex.QueryTake + "}}",
		dbflex.QuerySkip:  "WHERE ROWNUM <  {{." + dbflex.QuerySkip + "}}",
		dbflex.QueryGroup: "{{." + dbflex.QueryGroup + "}}",
		dbflex.QueryOrder: "ORDER BY {{." + dbflex.QueryOrder + "}}",
		dbflex.QueryInsert: "INSERT INTO {{." + dbflex.ConfigKeyTableName + "}} " +
			"({{.FIELDS}}) VALUES ({{.VALUES}})",
		dbflex.QueryUpdate: "UPDATE {{." + dbflex.ConfigKeyTableName + "}} " +
			"SET {{.FIELDVALUES}} {{." + dbflex.QueryWhere + "}}",
		dbflex.QueryDelete: "DELETE FROM {{." + dbflex.ConfigKeyTableName + "}} " +
			"{{." + dbflex.QueryWhere + "}}",
		dbflex.AggrMax:         "MAX({{.FIELD}})",
		dbflex.AggrMin:         "MIN({{.FIELD}})",
		string(dbflex.AggrSum): "SUM({{.FIELD}})",
		dbflex.AggrCount:       "COUNT(*)",
		dbflex.AggrAvg:         "AVG({{.FIELD}})",
	}
}

func (qr Query) ValueToSQlValue(v interface{}) string {
	if s, ok := v.(string); ok {
		//fmt.Println("datetime data: ", s)
		if dt, err := time.Parse(time.RFC3339, s); err == nil {
			return fmt.Sprintf("to_date('%s','yyyy-mm-dd hh24:mi:ss')", toolkit.Date2String(dt, "yyyy-MM-dd hh:mm:ss"))
		} else {
			return toolkit.Sprintf("'%s'", s)
		}
	} else if _, ok := v.(int); ok {
		return toolkit.Sprintf("%d", v)
	} else if _, ok = v.(float64); ok {
		return toolkit.Sprintf("%f", v)
	} else if _, ok = v.(time.Time); ok {
		dt := toolkit.Date2String(v.(time.Time), "yyyy-MM-dd hh:mm:ss")
		return fmt.Sprintf("to_date('%s','yyyy-mm-dd hh24:mi:ss')", dt)
	} else if b, ok := v.(bool); ok {
		if b {
			return "1"
		} else {
			return "0"
		}
	} else {
		vstr := toolkit.Sprintf("%v", v)
		if _, err := strconv.ParseFloat(vstr, 64); err == nil {
			return vstr
		} else {
			return "'" + vstr + "'"
		}
	}
}

/*
INSERT INTO TestModel (ID,Title,DataInt,DataDec,Created) VALUES ('data-0','Data title 0',3,30.700000,to_date('2019-02-23 06:45:21','yyyy-mm-dd hh24:mi:ss'))
*/
