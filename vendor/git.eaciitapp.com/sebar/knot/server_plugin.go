package knot

import (
	"strings"
)

// AddPlugin add server plugin
func (s *Server) AddPlugin(p Plugin) *Server {
	plugin := s.findPlugin(p.Name())

	if plugin != nil {
		return s
	}

	s.plugins = append(s.plugins, p)
	return s
}

// RemovePlugin remove server plugin
func (s *Server) RemovePlugin(name string) *Server {
	removedIdx := -1

	for idx, plugin := range s.plugins {
		if strings.ToLower(plugin.Name()) == strings.ToLower(name) {
			removedIdx = idx
			break
		}
	}

	if removedIdx != -1 {
		s.plugins = append(s.plugins[:removedIdx], s.plugins[removedIdx+1:]...)
	}

	return s
}

func (s *Server) findPlugin(name string) Plugin {
	for _, plugin := range s.plugins {
		if strings.ToLower(plugin.Name()) == strings.ToLower(name) {
			return plugin
		}
	}
	return nil
}
