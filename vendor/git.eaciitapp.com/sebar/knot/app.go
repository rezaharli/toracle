package knot

import (
	"html/template"
	"os"
	"reflect"
	"strings"

	"github.com/eaciit/toolkit"
)

// Application is a knot app
type Application struct {
	name    string
	plugins []Plugin
	routes  []*RouteItem

	viewsPath string
	funcs     template.FuncMap
	data      *SharedData
	statics   map[string]string

	sessions map[string]*SharedData
}

// NewApp create new application with default configuration
func NewApp() *Application {
	app := new(Application)
	app.plugins = []Plugin{}
	app.routes = []*RouteItem{}
	app.statics = map[string]string{}

	app.sessions = map[string]*SharedData{}
	return app
}

// SetViewsPath tell the knot app where is the view directory is
func (app *Application) SetViewsPath(p string) error {
	fi, err := os.Stat(p)
	if err != nil {
		return toolkit.Errorf("viewspath is invalid. %s. %s", p, err.Error())
	}

	if !fi.IsDir() {
		return toolkit.Errorf("viewspath is invalid. %s should be directory", p)
	}
	return nil
}

// Static add new static directory with given route,
// if route is exist then it will overwrite the older static directory
func (app *Application) Static(route string, dir string) *Application {
	app.statics[route] = dir
	return app
}

// Data return SharedData,
// If current SharedData is nil then new SharedData is initiated
func (app *Application) Data() *SharedData {
	if app.data == nil {
		app.data = NewSharedData()
	}
	return app.data
}

// AddRoute to application
func (app *Application) AddRoute(route string, fn func(*WebContext)) *RouteItem {
	ri := &RouteItem{Name: route, Pattern: route, Handler: fn}
	app.routes = append(app.routes, ri)
	return ri
}

// Register and object of struct (must be a struct) with given alias
// Then this function will scan for available knot handler (method of given struct that accept *knot.WebContext)
// If alias is empty string then it will use the lowercase of the struct name and remove the "controller" suffix if exist
func (app *Application) Register(obj interface{}, alias string) error {
	rv := reflect.ValueOf(obj)
	rt := rv.Type()

	name := alias
	if name == "" {
		name = strings.ToLower(reflect.Indirect(rv).Type().Name())
		if strings.HasSuffix(name, "controller") {
			name = name[:len(name)-len("controller")]
		}
	}

	routeItems := map[string]*RouteItem{}
	routeMethod := rv.MethodByName("Routes")
	if routeMethod.IsValid() {
		returnValues := routeMethod.Call([]reflect.Value{})
		rawRouteItems := returnValues[0].Interface().(map[string]*RouteItem)

		for k, v := range rawRouteItems {
			routeItems[strings.ToLower(k)] = v
		}
	}

	methodCount := rt.NumMethod()
	for i := 0; i < methodCount; i++ {
		mtd := rt.Method(i)
		tm := mtd.Type

		//--- check if it is a Handler
		isHandler := false
		if tm.NumIn() == 2 && tm.In(1).String() == "*knot.WebContext" {
			if tm.NumOut() == 0 {
				isHandler = true
			}
		}

		if isHandler {
			methodName := strings.ToLower(mtd.Name)
			pattern := strings.Join([]string{name, methodName}, "/")

			ri := new(RouteItem)
			if routeItem, ok := routeItems[methodName]; ok {
				ri = routeItem
			}

			ri.Name = pattern
			ri.Pattern = pattern
			ri.Handler = rv.MethodByName(mtd.Name).Interface().(func(*WebContext))

			app.routes = append(app.routes, ri)
		}
	}

	return nil
}
