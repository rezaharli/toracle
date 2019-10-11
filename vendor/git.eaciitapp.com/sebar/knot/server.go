package knot

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/eaciit/toolkit"
)

// Server is a knot server that HTTP/HTTPS
type Server struct {
	MultiApp bool

	routes          map[string]*RouteItem
	plugins         []Plugin
	proxySubDomains map[string]string

	address   string
	tlsConfig *tls.Config
	stop      chan bool
	errorTxt  chan string

	mux    *http.ServeMux
	hs     *http.Server
	status string
	logger *toolkit.LogEngine
	data   *SharedData

	apps              map[string]*Application
	viewsPath         string
	session           *sessionStore
	sessionCookieName string
}

// NewServer create new server with default configuration
func NewServer() *Server {
	s := new(Server)
	s.initRoute(true)
	s.initApps(true)
	s.initSessionStore(nil)
	return s
}

// Data get server SharedData
func (s *Server) Data() *SharedData {
	if s.data == nil {
		s.data = NewSharedData()
	}
	return s.data
}

func (s *Server) initApps(reset bool) *Server {
	if s.apps == nil || reset {
		s.apps = map[string]*Application{}
	}
	return s
}

// RegisterApp register knot app,
// If key is exist the it will overwrite the older one
func (s *Server) RegisterApp(app *Application, key string) *Server {
	s.initApps(false)
	key = strings.ToLower(key)
	app.name = key
	s.apps[key] = app
	return s
}

// SetTLS set tls.Config for this server
func (s *Server) SetTLS(config *tls.Config) *Server {
	s.tlsConfig = config
	return s
}

// Start server and listeing to request
func (s *Server) Start(address string) error {
	startStatus := make(chan string)

	if s.logger == nil {
		s.logger, _ = toolkit.NewLog(true, false, "", "", "")
	}

	s.logger.Infof("Initiate server for %s", address)
	s.address = address
	s.mux = http.NewServeMux()

	s.logger.Infof("registering %s", "/info")
	s.mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Knot Web Server 2.0"))
	})

	if err := s.prepareRoutes(); err != nil {
		return err
	}

	s.hs = &http.Server{Addr: address, Handler: s.mux,
		ErrorLog: log.New(os.Stdout, "http: ", log.LstdFlags)}

	if s.tlsConfig != nil {
		s.hs.TLSConfig = s.tlsConfig
	}

	var startTxt string
	s.stop = make(chan bool)
	go func() {
		go func() {
			time.Sleep(100 * time.Millisecond)
			if startTxt == "" {
				startStatus <- "OK"
			}
		}()

		s.status = "Running"
		var err error

		if s.tlsConfig != nil {
			err = s.hs.ListenAndServeTLS("", "")
		} else {
			err = s.hs.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			s.logger.Errorf("listen error. " + err.Error())
			startStatus <- err.Error()
		}
	}()

	go func() {
		<-s.stop

		s.session.close()
		if s.hs != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			err := s.hs.Shutdown(ctx)
			if err != nil {
				s.logger.Warningf("unable to stop server. %s", err.Error())
			} else {
				s.status = "Stop"
				s.logger.Info("Server is stopped")
			}
		} else {
			s.status = "Stop"
			s.logger.Info("Server is stopped")
		}
	}()

	startTxt = <-startStatus
	if startTxt != "OK" {
		return toolkit.Errorf(startTxt)
	}

	return nil
}

// Status return server status
func (s *Server) Status() string {
	return s.status
}

// SetViewsPath set views path for this server
func (s *Server) SetViewsPath(p string) {
	s.viewsPath = p
}

// Wait until the server status is actually stopped
func (s *Server) Wait() {
	for {
		<-time.After(100 * time.Millisecond)

		if s.status != "Running" {
			return
		}
	}
}

// Stop send stop server request
func (s *Server) Stop() {
	s.logger.Info("Server is requested to be stopped")
	s.stop <- true
}

// StopAndWait stop the server and wait until it accually stopped
func (s *Server) StopAndWait() {
	s.Stop()
	s.Wait()
}

// NoStdOut set log to blackhole
func (s *Server) NoStdOut() *Server {
	logger := s.Logger()
	logger.LogToStdOut = false
	s.SetLogger(logger)
	return s
}
