# scurvy
File synchronization... of sorts

[![Build Status](https://travis-ci.org/iggy/scurvy.svg?branch=master)](https://travis-ci.org/iggy/scurvy)

# What it does
* Accepts webhook input from various other pieces of software
* Publishes those events as msgs to Nats message queue
* Send slack/irc/etc messages based on MQ events

# Table of Contents
config/
  * common configuration code

input-webhook
  * Handles webhook input from sabnzbd, CouchPotato, and sickrage/sickbeard

irc/
  * IRC notifications (and future interface)

msgs/
  * common nats messaging code, shared message structs, etc

notify/
  * common notification code
  * mostly slack helpers/structs for now

notifyd/
  * daemon that listens on nats queues and sends notifications
  * currently only sends slack notifications

ping-listener/
  * Listens for pings from services and notifies (slack/etc) when a service hasn't been heard from  in a while

sendnatsmsg/
  * send test messages on nats queues
  * mostly used for development/debugging

stored/
  * db-api interface
  * all storage should go through this

syncd/
  * Small daemon that listens on nats queues and sync's files on new downloads
  * runs on anything down to a WD MyCloud
