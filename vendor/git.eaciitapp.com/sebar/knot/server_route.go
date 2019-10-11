package knot

import (
	"net/http"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/eaciit/toolkit"
)

func (s *Server) initRoute(reset bool) *Server {
	if s.routes == nil || reset {
		s.routes = map[string]*RouteItem{}
	}
	return s
}

// Route register a route enpoint for given uri and handler to server
func (s *Server) Route(uriPath string, handler func(*WebContext)) *RouteItem {
	s.initRoute(false)
	ri := new(RouteItem)
	ri.Name = uriPath
	ri.Pattern = uriPath
	ri.Handler = handler
	s.routes[uriPath] = ri
	return ri
}

// RouteStatic register static files route for given uri and file path
func (s *Server) RouteStatic(uriPath, filePath string) *RouteItem {
	s.initRoute(false)
	ri := new(RouteItem)
	ri.Name = uriPath
	ri.Pattern = uriPath
	ri.FilePath = filePath
	ri.IsStatic = true
	s.routes[uriPath] = ri
	return ri
}

func (s *Server) handleFunc(routeItem *RouteItem, app *Application) *Server {
	pattern := routeItem.Pattern
	fn := routeItem.Handler

	if s.mux != nil {
		if !strings.HasPrefix(pattern, "/") {
			pattern = "/" + pattern
		}

		plugins := append(s.plugins, routeItem.Plugins...)
		if app != nil {
			plugins = append(plugins, app.plugins...)
		}

		var unskippedPlugins []Plugin
		if len(routeItem.SkippedPluginsName) >= 0 {
			unskippedPlugins = []Plugin{}
			for _, plugin := range plugins {
				match := false
				for _, skippedPluginName := range routeItem.SkippedPluginsName {
					if plugin.Name() == skippedPluginName {
						match = true
						break
					}
				}

				if !match {
					unskippedPlugins = append(unskippedPlugins, plugin)
				}
			}
		} else {
			unskippedPlugins = plugins
		}

		fullFn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					s.Logger().Errorf("panic detected on %s: %v", r.URL.String(), rec)
				}
			}()

			ctx := newWebContext(s, app, w, r)
			ctx.Server().Logger().Infof("| Access | %s | %s", ctx.Request.URL, ctx.Request.RemoteAddr)

			if len(routeItem.AcceptedMethods) > 0 {
				if !toolkit.HasMember(routeItem.AcceptedMethods, r.Method) {
					ctx.Errorf("%s expect %v but called as %s", ctx.Request.URL, routeItem.AcceptedMethods, r.Method)
					return
				}
			}

			chainPlugins(unskippedPlugins, fn)(ctx)
		}

		s.logger.Infof("registering %s", pattern)
		s.mux.HandleFunc(pattern, fullFn)
	}
	return s
}

type reqexpRoute struct {
	app   *Application
	route *RouteItem
}

func (s *Server) prepareRoutes() error {
	isWin := runtime.GOOS == "windows"
	regexs := []*reqexpRoute{}

	s.initRoute(false)
	for _, item := range s.routes {
		if item.IsRegex {
			regexs = append(regexs, &reqexpRoute{nil, item})
			continue
		}

		if item.IsStatic {
			s.logger.Infof("registering %s = static for %s", item.Pattern, item.FilePath)
			s.mux.Handle(item.Pattern, http.StripPrefix(item.Pattern, http.FileServer(http.Dir(item.FilePath))))
		} else {
			s.handleFunc(item, nil)
		}
	}

	for appName, app := range s.apps {
		vp := app.viewsPath

		//-- app static
		for k, v := range app.statics {

			//-- path
			pathToStatic := ""
			if isWin {
				if strings.Contains(v, ":\\") {
					pathToStatic = v
				} else {
					pathToStatic = filepath.Join(vp, v)
				}
			} else {
				if strings.HasPrefix(v, "/") {
					pathToStatic = v
				} else {
					pathToStatic = filepath.Join(vp, v)
				}

				if !strings.HasPrefix(pathToStatic, "/") {
					pathToStatic = "/" + pathToStatic
				}
			}

			//-- pattern
			pattern := ""
			if s.MultiApp {
				pattern = appName
			}

			if strings.HasPrefix(k, "/") {
				pattern += k
			} else {
				pattern += "/"
				pattern += k
			}

			if !strings.HasPrefix(pattern, "/") {
				pattern = "/" + pattern
			}

			if !strings.HasSuffix(pattern, "/") {
				pattern += "/"
			}

			s.logger.Infof("registering %s = static for %s", pattern, pathToStatic)
			s.mux.Handle(pattern, http.StripPrefix(pattern, http.FileServer(http.Dir(pathToStatic))))
		}

		//-- app routes
		for _, item := range app.routes {
			if item.IsRegex {
				regexs = append(regexs, &reqexpRoute{app, item})
				continue
			}

			pattern := ""
			if s.MultiApp {
				pattern = appName
			}

			if strings.HasPrefix(item.Pattern, "/") {
				pattern += item.Pattern
			} else {
				pattern += "/"
				pattern += item.Pattern
			}

			if !strings.HasPrefix(pattern, "/") {
				pattern = "/" + pattern
			}

			s.handleFunc(item, app)
		}
	}

	//-- regex route, applicable only if no handle for /
	_, found := s.routes["/"]

	// if not found search for / handle in apps
	if !found {
		for _, app := range s.apps {
			for _, route := range app.routes {
				if route.Pattern == "/" {
					found = true
					break
				}
			}

			if !found {
				_, foundStatic := app.statics["/"]
				if foundStatic {
					found = true
					break
				}
			}

			if found {
				break
			}
		}
	}

	// If still not found add handle for /
	if !found {
		s.handleFunc(&RouteItem{Pattern: "/", Name: "/", Handler: func(ctx *WebContext) {
			for _, rei := range regexs {
				path := ctx.Request.URL.Path[1:]
				match, _ := regexp.MatchString(rei.route.Pattern, path)
				if match {
					ctx.app = rei.app
					rei.route.Handler(ctx)
					return
				}
			}

			http.Error(ctx.Writer,
				ctx.Request.URL.String()+" is not found | "+ctx.Request.URL.Host,
				http.StatusNotFound)
		}}, nil)
	}

	return nil
}
