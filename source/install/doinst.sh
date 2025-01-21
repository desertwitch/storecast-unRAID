#!/bin/bash
BOOT="/boot/config/plugins/dwstorecast"
DOCROOT="/usr/local/emhttp/plugins/dwstorecast"

chmod 755 /usr/bin/storecast
chmod 755 $DOCROOT/scripts/*
chmod 755 $DOCROOT/event/*

cp -n $DOCROOT/default.cfg $BOOT/dwstorecast.cfg >/dev/null 2>&1
cp -n /usr/local/emhttp/plugins/dwstorecast/misc/storecast.json /tmp/storecast.json
cp -n /usr/local/emhttp/plugins/dwstorecast/misc/storecast.json /tmp/storecast-dash.json

ln -sf /tmp/storecast.json /usr/local/emhttp/plugins/dwstorecast/storecast.json
ln -sf /tmp/storecast-dash.json /usr/local/emhttp/plugins/dwstorecast/storecast-dash.json
