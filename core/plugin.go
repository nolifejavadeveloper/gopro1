package core

var Plugins []Plugin

type Plugin struct {
	Name     string
	Init     func(p *Proxy) error
	Shutdown func(p *Proxy) error
}
