from golang:1.10 as build

workdir /go/src/go.jonnrb.io/mdns_repeater
copy . .
run go get ./cmd/mdns_repeater

from gcr.io/distroless/base
copy --from=build /go/bin/mdns_repeater /mdns_repeater
entrypoint ["/mdns_repeater"]
