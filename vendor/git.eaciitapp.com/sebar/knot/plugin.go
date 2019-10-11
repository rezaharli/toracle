package knot

// Plugin is a plugin or something similar like middleware
type Plugin interface {
	SetName(string) Plugin
	Name() string
	Handler(Handler) Handler
}

// PluginBase base struct to implement a plugin
type PluginBase struct {
	name string

	fn func(Handler) Handler
}

// SetName set plugin name
func (p *PluginBase) SetName(nm string) Plugin {
	p.name = nm
	return p
}

// Name get plugin name
func (p *PluginBase) Name() string {
	return p.name
}

// Handler get plugin handler
func (p *PluginBase) Handler(next Handler) Handler {
	if p.fn == nil {
		return next
	}

	return p.fn(next)
}

// NewPlugin initiate new plugin
func NewPlugin(name string, pluginFn func(h Handler) Handler) Plugin {
	p := new(PluginBase)
	p.SetName(name)
	p.fn = pluginFn
	return p
}

func chainPlugins(ps []Plugin, handler Handler) Handler {
	psCount := len(ps)
	if psCount == 0 {
		return handler
	} else if psCount == 1 {
		return ps[0].Handler(handler)
	}

	out := chainPlugins(ps[:psCount-1], ps[psCount-1].Handler(handler))
	return out
}
