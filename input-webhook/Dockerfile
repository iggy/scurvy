FROM scratch

COPY ddist/linux_amd64_input-webhook /input-webhook
COPY ddist/etc /

# just the one port that accepts webhook connections from sabnzbd/sickrage/CouchPotato
EXPOSE 38475

ENTRYPOINT ["/input-webhook"]
