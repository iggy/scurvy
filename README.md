# scurvy
File synchronization... of sorts

[![Build Status](https://travis-ci.org/iggy/scurvy.svg?branch=master)](https://travis-ci.org/iggy/scurvy)

## What it does
* Accepts webhook input from various other pieces of software
* Publishes those events as msgs to Nats message queue
* Send slack/irc/etc messages based on MQ events
* Sync files

## Architecture

             +-------------+   +---------------+
             | sendnatsmsg |   | input-webhook |
             +------+------+   +-------+-------+
                    |                  |
                    |                  |
                    +------v    v------+
                       +-----------+
                       |   gnats   |
                       +-+---+---+-+
                         |   |   |
                         |   |   |
             v-----------+   |   +--------v
        +-------+            v        +-----------+
        |  irc  |     +------+----+   |  syncd    |
        +-------+     |  notifyd  |   +-----------+
                      +-----------+

## Getting started

1. Install/configure gnats server (see example gnatsd.conf below)
1. Start input-webhook to receive messages from other software
1. Configure other software to send notifications to input-webhook
    * sabnzbd - configure -> notifications -> script = sabnzbd-notify.sh & parameters = json://127.0.0.1:38475/sabnzbd
    * sickrage/sickbeard/medusa - configure -> notifications -> KODI IP:Port = 127.0.0.1:38475
    * CouchPotato - configure -> notifications -> Webhook URL = http://localhost:38475/couchpotato
1. Start irc, notifyd on same server as nats/input-webhook for notifications
1. Run syncd on NAS device at remote location to sync files when they are downloaded

## Table of Contents
cmd/input-webhook/
* Handles webhook input from sabnzbd, CouchPotato, and sickrage/sickbeard/medusa

cmd/irc/
* IRC notifications (and future interface)

cmd/notifyd/
* daemon that listens on nats queues and sends notifications
* currently only sends slack notifications

cmd/sendnatsmsg/
* send test messages on nats queues
* mostly used for development/debugging

cmd/syncd/
* Small daemon that listens on nats queues and sync's files on new downloads
* runs on anything down to a WD MyCloud

pkg/config/
* common configuration code

pkg/errors/
* error handling boilerplate code

pkg/msgs/
* common nats messaging code, shared message structs, etc

pkg/notify/
* common notification code
* mostly slack helpers/structs for now

[TODO]
ping-listener/
* Listens for pings from services and notifies (slack/etc) when a service hasn't been heard from  in a while

stored/
* db-api interface
* all storage should go through this

## Examples

### gnatsd.conf

```json
port: 4242
http: 127.0.0.1:8282

tls {
  cert_file:  "/etc/letsencrypt/live/scurvy1.iggy.ninja-0001/fullchain.pem"
  key_file:   "/etc/letsencrypt/live/scurvy1.iggy.ninja-0001/privkey.pem"
  timeout:    2
  verify:     false
}

authorization {
  user: scurvy
  password: hunter2
  timeout: 0.5
}
```

## scurvy.yaml

```yaml
mq:
  tls: True
  host: natshost
  port: 4242
  user: scurvy
  password: hunter2

slack:
  webhook_key: MyVoiceIsMyPassport
  general_channel: '#general'
  admin_channel: '#admin'

syncd:
  newdownload:
    script: /path/to/scripts/sync.sh

```