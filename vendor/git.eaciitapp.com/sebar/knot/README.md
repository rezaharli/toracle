# knot

Knot 2.0

> `Knot 2.0` is not backward compatible with `knot.v1`

# Quick Start

Give me the code already, no documentation needed!
Ok.... here's the quickstart example

```go
package main

import (
	"net/http"

	"git.eaciitapp.com/sebar/knot"
	"github.com/eaciit/toolkit"
)

func main() {
	s := knot.NewServer()
	s.Route("hello", func(ctx *knot.WebContext) {
		ctx.Write([]byte("Awesome"), http.StatusOK)
	})

	s.Route("hola", func(ctx *knot.WebContext) {
		var o struct {
			FirstName, LastName string
		}

		ctx.GetPayload(&o)
		ctx.WriteJSON(toolkit.M{"FullName": o.FirstName + " " + o.LastName}, http.StatusOK)
	})

	s.Start(":8080")
	s.Wait()
}
```

# Features

Thos are some features of `knot 2.0` and also some improvement from `knot.v1`

* Flexible routing
* Sticky session 
* Session expiry 
* Plugin / midware
* Http2 support
* Http push
* Reverse proxy

## Flexible routing

You can do routing ethier in the `App` or directly in the `Server` object

```go
app := knot.NewApp()

app.AddRoute("hello", func(ctx *knot.WebContext) {
  ctx.Write([]byte("Hi!"), http.StatusOK)
})
```

```go
s := knot.NewServer()

s.Route("hello", func(ctx *knot.WebContext) {
  ctx.Write([]byte("Hi!"), http.StatusOK)
})
```

Or you can also do it the old way similar like `knot.v1`

```go

type Greeter struct {
}

func (g *Greeter) DoIt(ctx *knot.WebContext) {
	ctx.Write([]byte("Masoek pak eko"), http.StatusOK)
}

func main() {
	app := knot.NewApp()
	app.Register(Greeter{}, "")
	app.AddRoute("hello", func(ctx *knot.WebContext) {
		ctx.Write([]byte("Hi!"), http.StatusOK)
	})

	s := knot.NewServer()
	s.RegisterApp(app, "")
	s.Start(":8080")
	s.Wait()
}
```