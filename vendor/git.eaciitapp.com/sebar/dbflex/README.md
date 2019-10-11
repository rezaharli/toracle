# dbflex

**dbflex** is a new database library that meant to be a successor of **dbox**.

Detailed documentation can be found here [http://godoc.eaciitapp.com/pkg/git.eaciitapp.com/sebar/dbflex/](http://godoc.eaciitapp.com/pkg/git.eaciitapp.com/sebar/dbflex/)

## Driver supported by **dbflex**

* **MongoDB** (using mgo.v2)
* **MySQL** (using go-sql-driver/mysql)
* **Text** file (CSV Format)
* **JSON** file 

Upcoming driver

* **PostgreSQL**
* **ORACLE**

## Quick Start

You can find quickstart for each driver in their respective test file

Below is example quickstart for **JSON** driver

```go
package main

import (
	"fmt"
	"time"

	"git.eaciitapp.com/sebar/dbflex"
	"github.com/eaciit/toolkit"

	_ "git.eaciitapp.com/sebar/dbflex/drivers/json"
)

var (
	tableName = "test"
	workpath  = "/Users/sugab/sandbox/go/src/eaciit/try-dbflex"
)

func main() {
	conn, err := dbflex.NewConnectionFromURI(toolkit.Sprintf("json://localhost/%s?extension=json", workpath), toolkit.M{})
	if err != nil {
		panic(err)
	}

	err = conn.Connect()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	data := toolkit.M{
		"FullName": toolkit.M{
			"First": "Bagus",
			"Last":  "Cahyono",
		},
		"grade":    2,
		"_id":      "TEST-PARTIAL-M-1",
		"joinDate": time.Now(),
		"Role":     "Owner",
	}

	query, err := conn.Prepare(dbflex.From(tableName).Insert())
	if err != nil {
		panic(err)
	}

	_, err = query.Execute(toolkit.M{}.Set("data", data))
	if err != nil {
		panic(err)
	}

	buffer := []toolkit.M{}
	cmd := dbflex.From(tableName).Select().Where(dbflex.Eq("_id", data["_id"]))
	conn.Cursor(cmd, nil).Fetchs(&buffer, 0)

	fmt.Println(buffer)
}
```

> **Pro Tip**: This is only quickstart example, when writing real application throw the error instead of using panic.

## How to use

These are basic function available in dbflex. For more detailed API please read the go documentation.

### Create connection

```go
conn, err := dbflex.NewConnectionFromURI("mongodb://localhost/dbtest", nil)
// Check error here
```

### Connect

By default dbflex will not connect to the database, so you need to connect it explicitly. 

>You can use Connection Pooling to manage multiple connection, fo more infromation about connection pooling please scroll down ;)

```go
err = conn.Connect()
// Check error here

// Don't forget to close the connection after
defer conn.Close()
```

### Fetch

Getting single data with given command and will return error if no data match given command.

```go
emp := toolkit.M{}
cursor := conn.Cursor(
	dbflex.
		From(tableName).
		Where(dbflex.Eq("grade", 18)),
	nil)
err = cursor.Fetch(emp)
// Check error here

fmt.Println(emp)
```

### Fetchs

Getting multiple data with given command, if there is no data that match the given command, it will not return error but the buffer data length will be (no data returned).

```go
buffer := []toolkit.M{}
cursor := conn.Cursor(
	dbflex.
		From(tableName).
		Where(dbflex.Eq("grade", 18)).
		Aggr(dbflex.Avg("salary")).
		GroupBy(),
	nil)
err = cursor.Fetchs(&buffer, 0)
// Check error here

fmt.Println(buffer)
```

### Insert Data

Insert data accept either `struct` or `map[string]interface{}`. For some driver (text and json) it support inserting slice for much better performance when doing insert.

```go
data := toolkit.M{
	"Name": toolkit.M{
		"First": "Bagus",
		"Last":  "Cahyono",
	},
	"grade":    2,
	"_id":      "SGB-01",
	"joinDate": time.Now(),
	"Role":     "Developer",
}

query, err := conn.Prepare(dbflex.From(tableName).Insert())
// Check error here

_, err = query.Execute(toolkit.M{}.Set("data", data))
// Check error here too
```

## Update

Update require updated data to be supplied when calling execute command.

```go
data := toolkit.M{
	"_id":  "SGB-01",
	"Role": "Designer",
}

_, err = conn.Execute(
	dbflex.
		From(tableName).
		Where(dbflex.Eq("_id", data["_id"])).
		Update("Role"),
	toolkit.M{"data": data})
// Check error here
```

## Save

> Save currently only works for *MongoDB* driver 

Save data accept either `struct` or `map[string]interface{}`. Save will find the given `_id`. If the `_id` is exist it will update the existing value, if not it will insert new data.

```go
data := toolkit.M{
	"Name": toolkit.M{
		"First": "Bagus",
		"Last":  "Cahyono",
	},
	"grade":    2,
	"_id":      "SGB-01",
	"joinDate": time.Now(),
	"Role":     "Tester",
}

query, err := conn.Prepare(dbflex.From(tableName).Save())
// Check error here

_, err = query.Execute(toolkit.M{}.Set("data", data))
// Check error here too
```


## Delete

Delete will find data that match given filter then delete it.

> **CAREFUL** if no filter is given or the filter is `nil` then it will delete all data in other words truncate the table.

```go
err = conn.Execute(
	dbflex.
		From(tableName).
		Where(dbflex.Eq("_id", "SGB-01").
		Delete(), 
	nil)
// Check error here
```

## Commands

Below is list of available command in dbflex:

* **Reset** 
* **Select** 
* **From** 
* **Where** 
* **OrderBy** 
* **GroupBy** 
* **Aggr** 
* **Insert** 
* **Update** 
* **Delete** 
* **Save**  (MongoDB Only)
* **Take** 
* **Skip** 
* **Command** 
* **SQL** 


## Filters

Below is list of filters currently supported by dbflex:

* **And**
* **Or**
* **Eq**
* **Ne**
* **Gte**
* **Gt**
* **Lt**
* **Lte**
* **Range**
* **In**
* **Nin**
* **Contains**
* **StartWith**
* **EndWith**

## Aggregation

This library also support aggregation active record so you don't need to write pipe for simple aggregation.

Supported aggregation command

* **Sum**
* **Avg**
* **Min**
* **Max**
* **Count**

> For more complicated aggregation like push and unwind currently not yet supported by dbflex. So you still need to write your own pipe.

### Custom Pipe

Here how you can create custom pipe for mongo

```go
pipe := []toolkit.M{}
pipe = append(pipe, toolkit.M{
	"$match": toolkit.M{"grade": 1},
})

buffer := []toolkit.M{}
cursor := conn.Cursor(
	dbflex.
		From("employees").
		Command("pipe"), 
	toolkit.M{"pipe": pipe})
err = cursor.Fetchs(&buffer, 0)
// Check error here

fmt.Println(buffer)
```

# ORM

This library also include ORM helper, here how to use it.

First you need to create and connect the connection, asume that var `conn` is dbflex connection that already connected.

Then create a struct that implement `orm.DataModel`
 
```go
// orm.DataModel

type DataModel interface {
	TableName() string
	GetID() ([]string, []interface{})
	SetID([]interface{})

	PreSave(dbflex.IConnection)
	PostSave(dbflex.IConnection)
}
```

For example struct below is implementing all the Interface above.

```go
type UserModel struct {
	ID                string `json:"_id" bson:"_id" sql:"_id"`
	Name              string
	Title             string
	Grade             int
	LastUpdate        time.Time
	LastUpdateBy      string
}

func (m *UserModel) TableName() string {
	return "User"
}

func (m *UserModel) SetID(values []interface{}) {
	//-- do nothing
}

func (m *UserModel) GetID() ([]string, []interface{}) {
	return []string{"ID"}, []interface{}{m.ID}
}

func (d *DataModelBase) PreSave(conn dbflex.IConnection) {
	m.LastUpdate = time.Now()
	m.LastUpdateBy = "Admin"
}

func (d *DataModelBase) PostSave(conn dbflex.IConnection) {
	//-- do nothing
}
```

> The required method to be implemented actually only `GetID` and `TableName`. So for easier implementation you can embed `orm.DataModelBase` in your model.

Now you all set, and ready to go. Subsection below is the example what you can do with ORM.

### Get

Similar like `Fetch` it will get single data from given connection and data model. Data fetched injected directly to given model.

```go
user := new(UserModel)
err = orm.Get(conn, user)
// Check error here

fmt.Println(user)
```
### Gets

Similar like `Fetch` it will get single data from given connection and mode. Data fetched injected directly to given buffer.

```go
buffer := []toolkit.M{}
err = orm.Gets(conn, new(UserModel), &buffer, nil)

fmt.Println(buffer)
// Check error here
```

You also can add query parameter like this

```go
qry := dbflex.NewQueryParam()
qry.Where = dbflex.Gte("grade", 900)

buffer := []toolkit.M{}
err = orm.Gets(conn, new(UserModel), &buffer, qry)
// Check error here

fmt.Println(buffer)
```

### Insert

Insert new data from given model and connection

```go
newUser := new(UserModel)
newUser.ID = "EC"
newUser.Name = "EACIIT"
newUser.Title = "Pte. Ltd."
newUser.Grade = 7

err = orm.Insert(conn, newUser)
// Check error here
```

### Update

Update existing data

```go
newUser.Grade = 10
err = orm.Update(conn, newUser)
// Check error here
```

### Save

Save will check if there is existing data with given model, if true then it will update the data if not it will insert it as new data instead.

```go
newUser.Name = "EACIIT Vyasa"
err = orm.Save(conn, newUser)
// Check error here
```

### Delete 

Delete data from given model

```go
err = orm.Delete(conn, newUser)
// Check error here
```

# Connection Pooling

In this library it also include connection pooling

### Create New Pool

When creating new pool you need to specify the maximum connection can be hold and the function to establish the connection.

```go
connTxt := "mongodb://localhost/dbtest"

pooling := dbflex.NewDbPooling(10, func() (dbflex.IConnection, error) {
	conn, err := dbflex.NewConnectionFromURI(connTxt, nil)
	if err != nil {
		return nil, err
	}
	err = conn.Connect()
	if err != nil {
		return nil, err
	}
	return conn, nil
})

pooling.Timeout = 30 * time.Second

// don't forget to close the pool
defer pooling.Close()
```

You can also set `Timeout` so process will wait until given timeout before throwing error if no connection is given.

Default is `2 * time.Second`

### Get Connection

When you request a connection from dbpool it will check if any `PoolItem` is available.

If there is `PoolItem` available, it will return the `PoolItem`.

If no `PoolItem` is available and the current `PoolItem` number is still below maxium connection capacity, it will create a new one and then return it.

If the current `PoolItem` number is already maxed, then it will wait until timeout before throwing error.


```go
pconn, err := pooling.Get()
// Check error here

// Don't forget to release it
defer pconn.Release()

// Use the connection
orm.Save(pconn.Connection(), model)
```

# License

Copyright @ EACIIT Vyasa Pte. Ltd.