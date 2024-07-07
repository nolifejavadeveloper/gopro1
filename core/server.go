package core

import "net"

type Server struct {
}

type ServerInfo struct {
	Name string
	Addr net.Addr
}

func NewServerInfo(name string, addr net.Addr) *ServerInfo {
	return &ServerInfo{Name: name, Addr: addr}
}
