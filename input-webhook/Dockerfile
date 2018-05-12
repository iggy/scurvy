FROM scratch

COPY ../dist/linux_amd64_input-webhook /input-webhook

# scurvy looks for the config file in /etc by default
RUN mkdir /etc

# just the one port that accepts webhook connections from sabnzbd/sickrage/CouchPotato
EXPOSE 38475

ENTRYPOINT ["/input-webhook"]
