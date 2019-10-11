package knot

import (
	"net/http"
	"sync"
	"time"
)

type cookieStore struct {
	sync.RWMutex
	data map[string]*http.Cookie
}

// DefaultCookieExpire default duration that a cookie can live
var DefaultCookieExpire time.Duration

func (cs *cookieStore) initCookies(ctx *WebContext) {
	cs.Lock()
	if cs.data == nil {
		cs.data = make(map[string]*http.Cookie)
		cookies := ctx.Request.Cookies()
		for _, cookie := range cookies {
			cs.data[cookie.Name] = cookie
			http.SetCookie(ctx.Writer, cookie)
		}
	}
	cs.Unlock()
}

func (cs *cookieStore) getCookie(r *WebContext, name string, def string) (*http.Cookie, bool) {
	cs.initCookies(r)

	// first search on new cookies
	cs.RLock()
	c, exist := cs.data[name]
	cs.RUnlock()

	// when not found, try to search on request cookies
	if exist == false {
		var err error
		c, err = r.Request.Cookie(name)
		if err == nil {
			exist = true
		}
	}

	// when not exist and default is set
	// put cookie with default expire time
	if exist == false && len(def) > 0 {
		if int(DefaultCookieExpire) == 0 {
			DefaultCookieExpire = 30 * 24 * time.Hour
		}

		c = cs.setCookie(r, name, def, DefaultCookieExpire)
	}

	return c, exist
}

func (cs *cookieStore) setCookie(r *WebContext, name string, value string, expiresAfter time.Duration) *http.Cookie {
	cs.initCookies(r)

	c := &http.Cookie{}
	c.Name = name
	c.Value = value
	c.Path = "/"
	c.Expires = time.Now().Add(expiresAfter)
	c.Domain = r.Request.Host

	cs.Lock()
	cs.data[name] = c
	cs.Unlock()

	http.SetCookie(r.Writer, c)
	return c
}

func (cs *cookieStore) getAllCookies(ctx *WebContext) map[string]*http.Cookie {
	cs.initCookies(ctx)

	cs.RLock()
	data := cs.data
	cs.RUnlock()

	return data
}
