package knot

// RouteItem route item for each endpoint
type RouteItem struct {
	Name               string
	Pattern            string
	Handler            func(*WebContext)
	AcceptedMethods    []string
	IsRegex            bool
	SkippedPluginsName []string
	Plugins            []Plugin

	IsStatic bool
	FilePath string
}

// Methods define accepted method for given route
func (ri *RouteItem) Methods(ms ...string) *RouteItem {
	ri.AcceptedMethods = ms
	return ri
}

// SetPattern pattern for this item
func (ri *RouteItem) SetPattern(p string) *RouteItem {
	ri.Pattern = p
	return ri
}

// SetRegex for routing with regex
func (ri *RouteItem) SetRegex(r bool) *RouteItem {
	ri.IsRegex = r
	return ri
}

// SkipPlugins will skip plugin with given names
func (ri *RouteItem) SkipPlugins(names ...string) *RouteItem {
	ri.SkippedPluginsName = names
	return ri
}

// AddPlugins add specific plugins into this routes
func (ri *RouteItem) AddPlugins(plugins ...Plugin) *RouteItem {
	ri.Plugins = plugins
	return ri
}
