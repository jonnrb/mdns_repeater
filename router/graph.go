package router

import "net"

type graph map[net.PacketConn]map[net.PacketConn]struct{}

func (g graph) Add(src, dst net.PacketConn) {
	s := g[src]
	if s == nil {
		s = make(map[net.PacketConn]struct{})
		g[src] = s
	}
	s[dst] = struct{}{}
}
