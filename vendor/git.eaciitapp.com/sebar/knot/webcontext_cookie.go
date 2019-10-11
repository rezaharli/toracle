package knot

import (
	"net/http"
	"time"
)

// Cookies return all available cookies
func (ctx *WebContext) Cookies() map[string]*http.Cookie {
	return ctx.cookies.getAllCookies(ctx)
}

// SetCookie set cookie with given parameter
func (ctx *WebContext) SetCookie(name string, value string, duration time.Duration) {
	ctx.cookies.setCookie(ctx, name, value, duration)
}

// GetCookie get cookie with given name and default value if not exist
func (ctx *WebContext) GetCookie(name string, def string) *http.Cookie {
	out, _ := ctx.cookies.getCookie(ctx, name, def)
	return out
}
