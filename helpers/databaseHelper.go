package helpers

import (
	"errors"
	"log"
	"reflect"
	"strings"

	"github.com/gchaincl/dotsql"

	"github.com/eaciit/clit"
	_ "github.com/eaciit/gora"
	"github.com/eaciit/toolkit"

	"git.eaciitapp.com/sebar/dbflex"
	_ "git.eaciitapp.com/sebar/dbflex/drivers/mongodb"
)

var dbConnection dbflex.IConnection

func Database() dbflex.IConnection {
	if dbConnection != nil {
		return dbConnection
	}

	dbConf := clit.Config("default", "database", "").(map[string]interface{})

	log.Println("-> connecting to oracle database server")

	var err error
	dbConnection, err = NewOracleConnection(dbConf)
	if err != nil {
		log.Println("-> failed to connect to the oracle database server.", err.Error())
		// os.Exit(0)
	}

	return dbConnection
}

func NewOracleConnection(dbConf map[string]interface{}) (dbflex.IConnection, error) {
	connectionString := toolkit.Sprintf("oracle://%s:%s@%s:%s/%s", dbConf["username"], dbConf["password"].(string), dbConf["host"], dbConf["port"], dbConf["service"])
	conn, err := dbflex.NewConnectionFromURI(connectionString, nil)
	if err != nil {
		return nil, toolkit.Errorf("Unable to connect to the database server. %s", err.Error())
	}

	err = conn.Connect()
	if err != nil {
		return nil, toolkit.Errorf("Unable to connect to the database server. %s", err.Error())
	}

	return conn, nil
}

func TruncateSprintf(str string, args ...interface{}) (string, error) {
	n := strings.Count(str, "%s")
	if n > len(args) {
		return "", errors.New("Unexpected string:" + str)
	}
	return toolkit.Sprintf(str, args[:n]...), nil
}

func BuildQueryFromFile(filePath, queryName string, Colnames []string, args ...interface{}) (string, error) {
	dot, err := dotsql.LoadFromFile(filePath)
	if err != nil {
		return "", err
	}

	raw, err := dot.Raw(queryName)
	if err != nil {
		return "", err
	}

	for key, arg := range args {
		replacedArg := strings.ReplaceAll(toolkit.ToString(arg), "'", "''")
		args[key] = replacedArg
	}

	replaced := strings.ReplaceAll(raw, "%", "%%")
	replaced = strings.ReplaceAll(replaced, "?", "%s")

	return TruncateSprintf(replaced, args...)
}

type InsertParam struct {
	TableName       string
	Data            interface{}
	ContinueOnError bool
}

func Insert(param InsertParam) error {
	query, err := Database().Prepare(
		dbflex.From(param.TableName).Insert(),
	)
	if err != nil {
		return err
	}

	var lastError error

	switch reflect.TypeOf(param.Data).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(param.Data)

		for i := 0; i < s.Len(); i++ {
			_, err = query.Execute(toolkit.M{}.Set("data", s.Index(i).Interface()))
			lastError = err
			if !param.ContinueOnError && err != nil {
				return err
			}
		}
	default:
		_, err = query.Execute(toolkit.M{}.Set("data", param.Data))
		lastError = err
	}

	return lastError
}
