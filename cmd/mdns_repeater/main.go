/*
mdns_repeater forwards multicast mDNS messages across network boundaries,
allowing for discovery of devices that are L3 routable, but undiscoverable.

The device running this program must straddle the different networks involved.
The program considers a set of binary relationships between networks, where one
network is the "upstream" and one network is the "downstream". The "upstream"
network is the one that has devices that one wants to access on the "downstream"
network. These relationships are specified in yaml config file passed into the
program as an argument. The file looks like this:

  mappings:
    - upstream: eth0
      downstream: eth1
    - upstream: eth0
      downstream: eth2

The program recognizes network names by inteface name unless
-mapDockerNetworksToInterfaces is passed as a flag in which case the program
recognizes networks by their Docker names.
*/
package main // import "go.jonnrb.io/mdns_repeater/cmd/mdns_repeater"

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"
)

func openConfig() Config {
	name := flag.Arg(0)
	if name == "" {
		flag.Usage()
		os.Exit(1)
	}

	contents, err := ioutil.ReadFile(name)
	if err != nil {
		log.Fatalf("Could not open config file %q: %v", name, err)
	}

	cfg, err := ParseConfig(contents)
	if err != nil {
		log.Fatalf("Could not parse config file %q: %v", name, err)
	}

	return cfg
}

// If mapDockerNetworksToInterfaces is not set, network names are literal system
// interface names. Otherwise they correspond to Docker network names for
// networks connected to this container.
func provideResolveInterface() func(string) (*net.Interface, error) {
	if !*mapDockerNetworksToInterfaces {
		return net.InterfaceByName
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	netInfo, err := GetDockerNetworkInfo(ctx)
	if err != nil {
		log.Fatalln("Could not get network info:", err)
	}

	return func(dnet string) (*net.Interface, error) {
		name, err := netInfo.InterfaceForNetwork(dnet)
		if err != nil {
			return nil, err
		}
		return net.InterfaceByName(name)
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage %s config.yml:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	cfg := openConfig()

	r, err := cfg.Router(provideResolveInterface())
	if err != nil {
		log.Fatalln("Could not build mDNS router from config:", err)
	}

	r.Run(context.Background())
}
