package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"

	dockerTypes "github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/vishvananda/netlink"
)

var (
	mapDockerNetworksToInterfaces = flag.Bool("mapDockerNetworksToInterfaces", false, "The flag name is an open book")
)

type DockerNetworkInfo dockerTypes.NetworkSettings

func GetDockerNetworkInfo(ctx context.Context) (*DockerNetworkInfo, error) {
	cli, err := docker.NewEnvClient()
	if err != nil {
		return nil, err
	}

	hn, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	j, err := cli.ContainerInspect(ctx, hn)
	if err != nil {
		return nil, err
	}

	return (*DockerNetworkInfo)(j.NetworkSettings), nil
}

func (i DockerNetworkInfo) InterfaceForNetwork(dnet string) (string, error) {
	n, ok := i.Networks[dnet]
	if !ok {
		return "", fmt.Errorf("network %q not found on container info", dnet)
	}

	ip := net.ParseIP(n.IPAddress)
	if ip == nil {
		return "", fmt.Errorf("could not parse conatiner ip address %q", n.IPAddress)
	}

	l, err := linkForIP(ip)
	if err != nil {
		return "", err
	}

	return l.Attrs().Name, nil
}

func linkForIP(ip net.IP) (netlink.Link, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return nil, fmt.Errorf("error listing network links: %v", err)
	}

	for _, link := range links {
		addrs, err := netlink.AddrList(link, netlink.FAMILY_ALL)
		if err != nil {
			return nil, fmt.Errorf("error listing addrs on %q: %v", link.Attrs().Name, err)
		}
		for _, addr := range addrs {
			if addr.IPNet.IP.Equal(ip) {
				return link, nil
			}
		}
	}

	return nil, fmt.Errorf("could not find link for ip %v", ip)
}
