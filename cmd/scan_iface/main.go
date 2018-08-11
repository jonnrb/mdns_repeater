package main

import (
	"log"
	"net"
	"os"

	"github.com/miekg/dns"
	"go.jonnrb.io/mdns_repeater/mgrp"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Must provide interface as argument")
	}

	iface, err := net.InterfaceByName(os.Args[1])
	if err != nil {
		panic(err)
	}

	c, err := mgrp.New(net.IPv4(224, 0, 0, 251), 5353, *iface)
	if err != nil {
		panic(err)
	}

	log.Println("Scanning on", os.Args[1])
	var buf [1232]byte
	for {
		n, src, err := c.ReadFrom(buf[:])
		if err != nil {
			panic(err)
		}
		pkt := buf[:n]
		var msg dns.Msg
		if err := msg.Unpack(pkt); err != nil {
			panic(err)
		}
		log.Println(src, msg)
	}
}
