/*
Package router provides a declarative way of routing mDNS messages between
networks, possibly asymmetrically.
*/
package router // import "go.jonnrb.io/mdns_repeater/router"

import (
	"context"
	"net"

	"github.com/miekg/dns"
	"golang.org/x/sync/errgroup"
)

type R struct {
	qg, ag graph
	all    map[net.PacketConn]struct{}
}

// The router will make the upstream conn's resources flow downstream. This
// results in DNS questions flowing from the downstream network to the upstream
// one and DNS answers flowing back downstream.
//
// The router will close these conns on completion of Run.
//
func (r *R) Add(upstream, downstream net.PacketConn) {
	if r.all == nil {
		r.qg = make(graph)
		r.ag = make(graph)
		r.all = make(map[net.PacketConn]struct{})
	}

	r.qg.Add(upstream, downstream)
	r.ag.Add(downstream, upstream)

	r.all[upstream] = struct{}{}
	r.all[downstream] = struct{}{}
}

// Listens on added connections and routes messages accordingly.
func (r *R) Run(ctx context.Context) error {
	if r.all == nil {
		<-ctx.Done()
		return ctx.Err()
	}

	if d, ok := ctx.Deadline(); ok {
		for c := range r.all {
			c.SetDeadline(d)
		}
	}

	grp, ctx := errgroup.WithContext(ctx)

	go func() {
		<-ctx.Done()
		for c := range r.all {
			c.Close()
		}
	}()

	for c := range r.all {
		src := c
		grp.Go(func() error {
			var buf [1500]byte
			for {
				n, _, err := src.ReadFrom(buf[:])
				b := buf[:n]
				if err != nil {
					return err
				}

				var msg dns.Msg
				if err := msg.Unpack(b); err != nil {
					// TODO: Log this.
					continue // Ignore.
				}

				var mirrors []net.PacketConn
				if isQuestion := len(msg.Question) != 0; isQuestion {
					for c := range r.qg[src] {
						mirrors = append(mirrors, c)
					}
				} else {
					for c := range r.ag[src] {
						mirrors = append(mirrors, c)
					}
				}

				for _, m := range mirrors {
					// XXX: Only works with IPv4
					dst := net.UDPAddr{IP: net.IPv4(224, 0, 0, 251), Port: 5353}
					_, err := m.WriteTo(b, &dst)

					// TODO: Log this.
					_ = err
				}
			}
		})
	}

	return grp.Wait()
}
