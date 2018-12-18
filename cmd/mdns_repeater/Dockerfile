from quay.io/jonnrb/go as build
add . /src
run cd /src && CGO_ENABLED=0 go get ./cmd/mdns_repeater

from gcr.io/distroless/static
copy --from=build /go/bin/mdns_repeater /mdns_repeater
entrypoint ["/mdns_repeater"]
