package knot

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"
)

// GetValue is something that you should aware of
// C: Cookies
// H: Header
// Q: Query
// F: Form Value
// P: Payload (request.Body that have JSON content-type)
// S: Session
func (ctx *WebContext) GetValue(key string, defaultValue interface{}, sourcesOrder ...string) interface{} {
	sources := "CHQFSP"
	if len(sourcesOrder) > 0 {
		sources = strings.ToUpper(sourcesOrder[0])
	}

	funcMap := map[string]func(string, interface{}) interface{}{
		"C": ctx.getCookies,
		"H": ctx.getHeader,
		"Q": ctx.getQuery,
		"F": ctx.getFormValue,
		"S": ctx.getSession,
		"P": ctx.getPayload,
	}

	for _, source := range sources {
		if f, ok := funcMap[string(source)]; ok {
			data := f(key, defaultValue)
			if data != defaultValue {
				return data
			}
		}
	}

	return defaultValue
}

func (ctx *WebContext) getQuery(key string, defaultValue interface{}) interface{} {
	data := ctx.Request.URL.Query().Get(key)
	if data == "" {
		return defaultValue
	}

	return data
}

func (ctx *WebContext) getHeader(key string, defaultValue interface{}) interface{} {
	data := ctx.Request.Header.Get(key)
	if data == "" {
		return defaultValue
	}

	return data
}

func (ctx *WebContext) getCookies(key string, defaultValue interface{}) interface{} {
	data := ctx.GetCookie(key, "")
	if data == nil {
		return defaultValue
	}

	return data.Value
}

func (ctx *WebContext) getFormValue(key string, defaultValue interface{}) interface{} {
	ctx.Request.ParseForm()
	data := ctx.Request.FormValue(key)
	if data == "" {
		return defaultValue
	}

	return data
}

func (ctx *WebContext) getSession(key string, defaultValue interface{}) interface{} {
	return ctx.GetSession(key, defaultValue)
}

func (ctx *WebContext) getPayload(key string, defaultValue interface{}) interface{} {
	if strings.Contains(ctx.Request.Header.Get("Content-Type"), "application/json") {
		payload := map[string]interface{}{}

		b := bytes.NewBuffer(make([]byte, 0))
		reader := io.TeeReader(ctx.Request.Body, b)

		err := json.NewDecoder(reader).Decode(&payload)
		if err != nil {
			return defaultValue
		}
		defer ctx.Request.Body.Close()
		ctx.Request.Body = ioutil.NopCloser(b)

		if data, ok := payload[key]; ok {
			return data
		}
	}

	return defaultValue
}
