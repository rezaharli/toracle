package knot

import (
	"strings"
)

// AddPlugin adding plugin to knot application
func (a *Application) AddPlugin(p Plugin) *Application {
	plugin := a.findPlugin(p.Name())

	if plugin != nil {
		return a
	}

	a.plugins = append(a.plugins, p)
	return a
}

// RemovePlugin remove plugin from knot application
func (a *Application) RemovePlugin(name string) *Application {
	removedIdx := -1
	lowerName := strings.ToLower(name)

	for idx, plugin := range a.plugins {
		if strings.ToLower(plugin.Name()) == lowerName {
			removedIdx = idx
			break
		}
	}

	if removedIdx != -1 {
		a.plugins = append(a.plugins[:removedIdx], a.plugins[removedIdx+1:]...)
	}

	return a
}

func (a *Application) findPlugin(name string) Plugin {
	lowerName := strings.ToLower(name)
	for _, plugin := range a.plugins {
		if strings.ToLower(plugin.Name()) == lowerName {
			return plugin
		}
	}
	return nil
}
