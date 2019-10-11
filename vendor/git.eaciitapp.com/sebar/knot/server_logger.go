package knot

import "github.com/eaciit/toolkit"

// SetLogger set logger that use toolkit.LogEngine
func (s *Server) SetLogger(l *toolkit.LogEngine) *Server {
	s.logger = l
	return s
}

// Logger get looger,
// If logger is nil then return newly initiate toolkit.LogEngine
func (s *Server) Logger() *toolkit.LogEngine {
	if s.logger == nil {
		s.logger, _ = toolkit.NewLog(true, false, "", "", "")
	}
	return s.logger
}
