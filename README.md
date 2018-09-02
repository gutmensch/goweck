# GoWeck - Alarm Clock for Raumserver

1. Load streams or playlists on alarm

2. Increase volume over time

3. Simple frontend to create/edit/delete alarms


# Installation

0. you need node-raumserver running somewhere

1. install golang package, set GOPATH to $HOME/go

2. install mongodb-server package

3. execute go get github.com/gutmensch/goweck, go install

4. create systemd file for service like this

```root@raspi3:/home/pi/go# cat /etc/systemd/system/goweck.service
[Unit]
Description=GoWeck Radio Wecker
After=network.target mongodb.service

[Service]
Type=simple
Environment=TZ=Europe/Berlin
Environment=MONGODB_URI=mongodb://127.0.0.1:27017
Environment=MONGODB_DATABASE=goweck
Environment=MONGODB_DROP=false
Environment=RAUMSERVER_URI=http://127.0.0.1:3535/raumserver
Environment=RAUMSERVER_ZONE=uuid:C43C1A1D-AED1-472B-B0D0-210B7925000E
Environment=PUSHOVER_APP_TOKEN=foobar
Environment=PUSHOVER_USER_TOKEN=barfoo
Environment=RADIO_CHANNEL=http://mp3channels.webradio.rockantenne.de/alternative
Environment=DEBUG=false
Environment=RAUMSERVER_DEBUG=false
Environment=LISTEN=:8081
Restart=always
RestartSec=3
ExecStart=/home/pi/go/bin/goweck

[Install]
WantedBy=multi-user.target
```
