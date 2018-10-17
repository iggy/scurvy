# Run tests stage
FROM golang:alpine as test

WORKDIR /go/src/github.com/iggy/scurvy/

RUN apk add git upx gcc libc-dev

RUN go get -u golang.org/x/lint/golint \
	honnef.co/go/tools/cmd/megacheck \
	github.com/fzipp/gocyclo

# Use add here to invalidate the cache
ADD . /go/src/github.com/iggy/scurvy/

# install deps the easy way
RUN go get github.com/iggy/scurvy/...

# These are all separate so failures are a little easier to track
RUN gofmt -l -s -w ./cmd ./pkg
# RUN test -z $(gofmt -s -l $GO_FILES)
# go test -race basically doesn't work with alpine/musl
# RUN go test -v -race ./...
RUN go vet ./...
RUN megacheck ./...
# RUN gocyclo -over 19 $GO_FILES
RUN golint -set_exit_status $(go list ./...)



# Build binaries stage
FROM golang:alpine as build

WORKDIR /go/src/github.com/iggy/scurvy/

RUN apk add git upx gcc libc-dev

RUN go get -u github.com/mitchellh/gox \
	github.com/tcnksm/ghr

# Use add here to invalidate the cache
ADD . /go/src/github.com/iggy/scurvy/

# install deps the easy way
RUN go get github.com/iggy/scurvy/...

RUN gox -arch="amd64 arm 386" -os="linux" -output="dist/{{.OS}}_{{.Arch}}_{{.Dir}}" -ldflags='-extldflags "-static" -s -w' -tags='netgo' ./...
RUN mkdir -p /ddist/etc
# we only need to build the
RUN CGO_ENABLED=0 gox -arch="amd64" -os="linux" -output="/ddist/{{.OS}}_{{.Arch}}_{{.Dir}}" -ldflags='-extldflags "-static" -s -w' -tags='netgo' ./...
RUN upx /ddist/linux*



# This builds the irc image from build binaries stage output
FROM scratch as irc
COPY --from=build /ddist/linux_amd64_irc /ircbot
COPY --from=build /ddist/etc /
ENTRYPOINT ["/ircbot"]



# This builds the notifyd image from build binaries stage output
FROM scratch as notifyd
COPY --from=build /ddist/linux_amd64_notifyd /notifyd
COPY --from=build /ddist/etc /
ENTRYPOINT ["/notifyd"]



# This builds the input-webhook image from build binaries stage output
FROM scratch as input-webhook
COPY --from=build /ddist/linux_amd64_input-webhook /input-webhook
COPY --from=build /ddist/etc /
# just the one port that accepts webhook connections from sabnzbd/sickrage/CouchPotato
EXPOSE 38475
ENTRYPOINT ["/ircbot"]
