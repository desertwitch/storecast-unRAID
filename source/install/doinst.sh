#!/bin/bash
cp -n /usr/local/emhttp/plugins/dwstorecast/storecast-def.json /tmp/storecast.json
cp -n /usr/local/emhttp/plugins/dwstorecast/storecast-def.json /tmp/storecast-dash.json
ln -sf /tmp/storecast.json /usr/local/emhttp/plugins/dwstorecast/storecast.json
ln -sf /tmp/storecast-dash.json /usr/local/emhttp/plugins/dwstorecast/storecast-dash.json
