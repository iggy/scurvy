FROM golang:alpine3.12 as build

WORKDIR /src

COPY go.* /src/

RUN go mod download

COPY . /src/

RUN mkdir bins/

RUN go build -tags netgo -ldflags='-extldflags="-static" -s -w' -o bins/ ./...

# TODO add upx back

# This builds the irc image from build binaries stage output
FROM scratch as irc
COPY --from=build /src/bins/irc /ircbot
COPY --from=build /etc/ssl /etc
ENTRYPOINT ["/ircbot"]



# This builds the notifyd image from build binaries stage output
FROM scratch as notifyd
COPY --from=build /src/bins/notifyd /notifyd
COPY --from=build /etc/ssl /etc/
ENTRYPOINT ["/notifyd"]



# This builds the input-webhook image from build binaries stage output
FROM scratch as input-webhook
COPY --from=build /src/bins/input-webhook /input-webhook
COPY --from=build /etc/ssl /etc/
# just the one port that accepts webhook connections from sabnzbd/sickrage/CouchPotato
EXPOSE 38475
ENTRYPOINT ["/input-webhook"]



# This builds the syncd image from build binaries stage output
# syncd runs a shell script to do the actual downloading, so can't use `scratch`
FROM alpine:3.20.0 as syncd
COPY --from=build /src/bins/syncd /syncd
COPY --from=build /etc/ssl /etc/
# COPY --from=build /go/src/github.com/iggy/scurvy/cmd/syncd/sync_files.sh /

# Need the ca-certificates for the NATS TLS cert and using rsync for the
# file sync for now
RUN apk --no-cache add rsync ca-certificates

# ENV SCURVY_BASE_URL "https://scurvy"
# ENV SCURVY_DL_DIR "/scurvy/"
# ENV SCURVY_COMPLETE_URL "scurvy/complete"
# Where all the files are stored
# VOLUME ["/scurvy"]
# The script to run when syncd gets a new download message
# If using rsync/ssh/etc, you'll need to also pass in ssh keys
# VOLUME ["/sync_files.sh"]
ENTRYPOINT ["/syncd"]
