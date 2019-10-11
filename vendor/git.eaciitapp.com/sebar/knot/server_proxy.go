package knot

import (
	"log"
	"net/http/httputil"
	"net/url"
	"strings"
)

// ReverseProxyType type of reverse proxy
type ReverseProxyType string

const (
	// SubDomain type reverse proxy
	SubDomain ReverseProxyType = "SubDomain"
	// VirtualDirectory type reverse proxy
	VirtualDirectory ReverseProxyType = "VirtualDirectory"
)

// ReverseProxy register a proxy either sub domain proxy or virtual directory proxy
func (s *Server) ReverseProxy(pattern, target string, proxyType ReverseProxyType) {
	if proxyType == VirtualDirectory {
		rpHandler := func(ctx *WebContext) {
			if pattern[0] != '/' {
				pattern = "/" + pattern
			}

			url, _ := url.Parse(target)
			if strings.HasPrefix(ctx.Request.RequestURI, pattern) && ctx.Request.RequestURI != pattern {
				ctx.Request.RequestURI = ctx.Request.RequestURI[len(pattern):]
				ctx.Request.URL.Path = strings.Split(ctx.Request.RequestURI, "?")[0]
			}

			proxy := httputil.NewSingleHostReverseProxy(url)
			ctx.Writer.Header().Set("X-ReverseProxy", "Knot")
			ctx.Writer.Header().Set("X-Forwarded-Host", ctx.Request.Header.Get("Host"))
			proxy.ServeHTTP(ctx.Writer, ctx.Request)
		}

		s.Route(pattern, rpHandler)
		s.Route(pattern+"/([\\w\\W]+)", rpHandler).SetRegex(true)
	} else {
		if len(s.proxySubDomains) == 0 {
			s.proxySubDomains = map[string]string{}

			s.AddPlugin(NewPlugin("knot_proxy", func(next Handler) Handler {
				return func(ctx *WebContext) {
					host := ctx.Request.Host

					if target, ok := s.proxySubDomains[host]; ok {
						url, err := url.Parse(target)
						if err != nil {
							log.Println("target parse fail:", err)
							return
						}

						proxy := httputil.NewSingleHostReverseProxy(url)
						ctx.Writer.Header().Set("X-ReverseProxy", "Knot")
						ctx.Writer.Header().Set("X-Forwarded-Host", ctx.Request.Header.Get("Host"))
						proxy.ServeHTTP(ctx.Writer, ctx.Request)
						return
					}
				}
			}))
		}

		s.proxySubDomains[pattern] = target
	}
}
