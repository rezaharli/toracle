package knot

import (
	"time"

	"git.eaciitapp.com/sebar/dbflex"
)

func (s *Server) initSessionStore(conn dbflex.IConnection) {
	s.sessionCookieName = "ECSESSIONID"
	ss := newSessionStore()
	if conn != nil {
		ss.connection = conn
	}
	s.session = ss
}

// SetSessionConnection set session connection,
// only support dbflex.IConnection
func (s *Server) SetSessionConnection(conn dbflex.IConnection) *Server {
	s.initSessionStore(conn)
	return s
}

// SetSessionCookieName set session cookie name
func (s *Server) SetSessionCookieName(name string) *Server {
	s.sessionCookieName = name
	return s
}

// SetSessionDuration set how long session can live
func (s *Server) SetSessionDuration(d time.Duration) *Server {
	if s.session == nil {
		s.session = newSessionStore()
	}
	s.session.setSessionDuration(d)
	return s
}
