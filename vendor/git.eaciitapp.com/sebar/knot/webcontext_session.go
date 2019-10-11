package knot

import "github.com/eaciit/toolkit"

// SetSession set seesion with given key and value
func (ctx *WebContext) SetSession(key string, value interface{}) error {
	sessionCookie, _ := ctx.cookies.getCookie(ctx,
		ctx.server.sessionCookieName, toolkit.RandomString(64))
	return ctx.server.session.set(sessionCookie.Value, key, value)
}

// GetSession get session with given key and default value if not exist
func (ctx *WebContext) GetSession(key string, def interface{}) interface{} {
	sessionCookie, exist := ctx.cookies.getCookie(ctx,
		ctx.server.sessionCookieName, toolkit.RandomString(64))
	if !exist {
		return def
	}
	out := ctx.server.session.get(sessionCookie.Value, key, def)
	return out
}
