# scurvy
File synchronization... of sorts

[![Build Status](https://travis-ci.org/iggy/scurvy.svg?branch=master)](https://travis-ci.org/iggy/scurvy)

# What it does
* Accepts webhook input from various other pieces of software
* Publishes those events as msgs to Nats message queue
* Send slack/irc/etc messages based on MQ events
