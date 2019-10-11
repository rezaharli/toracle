package knot

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"html/template"

	"github.com/eaciit/toolkit"
)

// WebContext holds the context of current request
type WebContext struct {
	Ctx     *context.Context
	Request *http.Request
	Writer  http.ResponseWriter
	Value   map[string]interface{}

	app            *Application
	server         *Server
	controllerName string
	cookies        *cookieStore
}

func newWebContext(
	s *Server, a *Application,
	w http.ResponseWriter, r *http.Request) *WebContext {
	ctx := new(WebContext)
	ctx.server = s
	ctx.app = a
	ctx.Writer = w
	ctx.Request = r
	ctx.Value = map[string]interface{}{}
	ctx.cookies = new(cookieStore)
	ctx.cookies.initCookies(ctx)
	return ctx
}

// GetPayload get the request body, assume that the request is json type,
// decode it, and put it in given output parameter
func (ctx *WebContext) GetPayload(output interface{}) error {
	if ctx.Request == nil {
		return toolkit.Errorf("Request is nil")
	}

	decoder := json.NewDecoder(ctx.Request.Body)
	defer ctx.Request.Body.Close()

	err := decoder.Decode(output)

	if err != nil {
		return toolkit.Errorf("payload error, decoding payload. %s", err.Error())
	}

	return nil
}

// Errorf write error response with response code http.StatusInternalServerError
func (ctx *WebContext) Errorf(pattern string, parm ...interface{}) error {
	if ctx.Writer == nil {
		return toolkit.Errorf("ResponseWriter is nil")
	}

	errorTxt := toolkit.Sprintf(pattern, parm...)
	ctx.Write([]byte(errorTxt), http.StatusInternalServerError)
	return toolkit.Error(errorTxt)
}

// WriteJSON encode given data to JSON and then write it into the writer
func (ctx *WebContext) WriteJSON(data interface{}, status int) error {
	w := ctx.Writer
	if w == nil {
		return toolkit.Error("Writer is nil")
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
	return nil
}

// Server return current server
func (ctx *WebContext) Server() *Server {
	return ctx.server
}

// App return current app
func (ctx *WebContext) App() *Application {
	return ctx.app
}

// WriteTemplate write given data with given template name
func (ctx *WebContext) WriteTemplate(data interface{}, templateName string) error {
	isWin := runtime.GOOS == "windows"

	if ctx.Writer == nil {
		return toolkit.Error("Writer is nil")
	}

	//-- get template
	appName := ""
	if ctx.app != nil {
		appName = ctx.app.name
	}

	//-- get template location
	pathToTemplate := ""
	if isWin && strings.Contains(templateName, ":\\") {
		pathToTemplate = templateName
	} else if strings.HasPrefix(templateName, "/") {
		pathToTemplate = templateName
	} else {
		if ctx.app == nil {
			pathToTemplate = ctx.server.viewsPath
		} else {
			pathToTemplate = ctx.app.viewsPath
		}

		if templateName == "" {
			uris := strings.Split(ctx.Request.URL.Path, "/")
			controllerAction := uris[len(uris)-1]
			pathToTemplate = filepath.Join(pathToTemplate, ctx.controllerName, controllerAction+".html")
		} else {
			pathToTemplate = filepath.Join(pathToTemplate, ctx.controllerName, templateName+".html")
		}
	}

	ft, err := os.Open(pathToTemplate)
	if err != nil {
		if ctx.app != nil {
			return ctx.Errorf("app:%s template:%s is not exist", ctx.app.name, templateName)
		}

		return ctx.Errorf("template:%s is not exist", templateName)
	}
	defer ft.Close()

	htmlTxtBytes, err := ioutil.ReadAll(ft)
	if err != nil {
		return ctx.Errorf("unable to read %s. %s", pathToTemplate, err.Error())
	}

	//--- funcs
	tfuncs := template.FuncMap{
		"BaseUrl": func() string {
			base := "/"
			if appName != "" {
				base += strings.ToLower(appName)
			}

			if base != "/" {
				base += "/"
			}

			return base
		},
		"UnescapeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"NoCacheUrl": func(s string) string {
			concatenator := "?"
			if strings.Contains(s, "?") {
				concatenator = `&`
			}

			randomString := toolkit.RandomString(32)
			noCachedURL := fmt.Sprintf("%s%snocache=%s", s, concatenator, randomString)
			return noCachedURL
		}}

	//--- app funcs
	if ctx.app != nil && ctx.app.funcs != nil {
		for k, f := range ctx.app.funcs {
			tfuncs[appName+"_"+k] = f
		}
	}

	//--- include files
	t, err := template.New("main").Funcs(tfuncs).Parse(string(htmlTxtBytes))
	if err != nil {
		return ctx.Errorf("unable to parse template. %s", err.Error())
	}

	err = t.Execute(ctx.Writer, data)
	if err != nil {
		return ctx.Errorf("unable to execute template. %s", err.Error())
	}

	return nil
}

func (ctx *WebContext) Write(data []byte, status int) error {
	if ctx.Writer == nil {
		return toolkit.Error("Writer is nil")
	}

	ctx.Writer.WriteHeader(status)
	_, err := ctx.Writer.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// GetMultiPart get multipart form data for current request
func (ctx *WebContext) GetMultiPart() (map[string][]*multipart.FileHeader, map[string][]string, error) {
	var e error
	if ctx.Request == nil {
		return nil, nil, toolkit.Error("request is nil")
	}
	e = ctx.Request.ParseMultipartForm(1024 * 1024 * 1024 * 1024)
	if e != nil {
		return nil, nil, toolkit.Errorf("unable to parse: %s", e.Error())
	}

	m := ctx.Request.MultipartForm
	return m.File, m.Value, nil
}
