package main

import (
	"net"

	"go.jonnrb.io/mdns_repeater/mgrp"
	"go.jonnrb.io/mdns_repeater/router"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Mappings []Mapping `yaml:"mappings"`
}

type Mapping struct {
	Upstream   string `yaml:"upstream"`
	Downstream string `yaml:"downstream"`
}

func ParseConfig(contents []byte) (Config, error) {
	var c Config
	err := yaml.Unmarshal(contents, &c)
	return c, err
}

func (c Config) Router(resolveInterface func(string) (*net.Interface, error)) (*router.R, error) {
	var r router.R

	getConn := func() func(name string) (net.PacketConn, error) {
		m := make(map[string]net.PacketConn)
		return func(name string) (net.PacketConn, error) {
			if conn, ok := m[name]; ok {
				return conn, nil
			}
			iface, err := resolveInterface(name)
			if err != nil {
				return nil, err
			}
			conn, err := mgrp.New(net.IPv4(224, 0, 0, 251), 5353, *iface)
			if err != nil {
				return nil, err
			}
			m[name] = conn
			return conn, nil
		}
	}()

	for _, m := range c.Mappings {
		upstream, err := getConn(m.Upstream)
		if err != nil {
			return nil, err
		}
		downstream, err := getConn(m.Downstream)
		if err != nil {
			return nil, err
		}

		r.Add(upstream, downstream)
	}

	return &r, nil
}
