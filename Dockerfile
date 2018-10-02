FROM golang:alpine as test

RUN apk add git upx

RUN go get -u github.com/golang/lint/golint \
	honnef.co/go/tools/cmd/megacheck \
	github.com/fzipp/gocyclo \
	github.com/mitchellh/gox \
	github.com/tcnksm/ghr

RUN gofmt -l -s -w ./irc ./cmd ./pkg
# RUN test -z $(gofmt -s -l $GO_FILES)
RUN go test -v -race ./...
RUN go vet ./...
RUN megacheck ./...
# RUN gocyclo -over 19 $GO_FILES
RUN golint -set_exit_status $(go list ./...)

FROM golang:alpine as build

RUN gox -arch="amd64 arm 386" -os="linux" -output="dist/{{.OS}}_{{.Arch}}_{{.Dir}}" -ldflags='-extldflags "-static" -s -w' -tags='netgo' ./...
RUN mkdir -p /ddist/etc
RUN CGO_ENABLED=0 gox -arch="amd64" -os="linux" -output="/ddist/{{.OS}}_{{.Arch}}_{{.Dir}}" -ldflags='-extldflags "-static" -s -w' -tags='netgo' ./...
RUN upx /ddist/linux*

# This builds the irc image from above stage output
FROM scratch as irc
COPY --from=build /ddist/linux_amd64_irc /ircbot
COPY --from=build /ddist/etc /
ENTRYPOINT ["/ircbot"]

# This builds the input-webhook image from above stage output
FROM scratch as input-webhook
COPY --from=build /ddist/linux_amd64_input-webhook /input-webhook
COPY --from=build /ddist/etc /
# just the one port that accepts webhook connections from sabnzbd/sickrage/CouchPotato
EXPOSE 38475
ENTRYPOINT ["/ircbot"]
