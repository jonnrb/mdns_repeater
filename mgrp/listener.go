/*
Package mgrp implements interface-honed multicast UDP.

The Linux multicast API lets you join a multicast group on a specific interface,
but doesn't disambiguate received traffic based on the receiving interface,
leading to unexpected consequences in applications that care about the source of
the traffic.
*/
package mgrp // import "go.jonnrb.io/mdns_repeater/mgrp"

import (
	"net"

	"golang.org/x/net/ipv4"
)

type conn struct {
	*ipv4.PacketConn

	ip    net.IP
	iface net.Interface
}

// Joins the multicast group specified by ip on iface and listens on the
// specified port for UDP traffic. Incoming traffic is filtered to packets sent
// specifically to this multicast group and received on iface.
func New(ip net.IP, port int, iface net.Interface) (*conn, error) {
	udp, err := net.ListenUDP("udp", &net.UDPAddr{IP: ip, Port: port})
	if err != nil {
		return nil, err
	}

	p := ipv4.NewPacketConn(udp)
	if err := p.JoinGroup(&iface, &net.UDPAddr{IP: ip}); err != nil {
		udp.Close()
		return nil, err
	}

	if err := p.SetControlMessage(ipv4.FlagDst|ipv4.FlagInterface, true); err != nil {
		p.LeaveGroup(&iface, &net.UDPAddr{IP: ip})
		udp.Close()
		return nil, err
	}

	return &conn{
		PacketConn: p,
		ip:         ip,
		iface:      iface,
	}, nil
}

func (c *conn) Close() error {
	if err := c.PacketConn.LeaveGroup(&c.iface, &net.UDPAddr{IP: c.ip}); err != nil {
		return err
	}
	return c.PacketConn.Close()
}

// Reads incoming UDP packets in a loop until one matches the multicast group
// and interface passed into New.
func (c *conn) ReadFrom(b []byte) (n int, src net.Addr, err error) {
	var cm *ipv4.ControlMessage
	for {
		n, cm, src, err = c.PacketConn.ReadFrom(b)
		if err != nil {
			return
		}

		if !cm.Dst.Equal(c.ip) {
			continue
		} else if cm.IfIndex != c.iface.Index {
			continue
		} else {
			return
		}
	}
}

// Writes a UDP packet to dst, setting the outgoing interface to the one
// specified in New.
func (c *conn) WriteTo(b []byte, dst net.Addr) (int, error) {
	cm := ipv4.ControlMessage{IfIndex: c.iface.Index}
	return c.PacketConn.WriteTo(b, &cm, dst)
}
