#!/bin/bash
#
# Copyright Derek Macias (parts of code from NUT package)
# Copyright macester (parts of code from NUT package)
# Copyright gfjardim (parts of code from NUT package)
# Copyright SimonF (parts of code from NUT package)
# Copyright Lime Technology (any and all other parts of Unraid)
#
# Copyright desertwitch (as author and maintainer of this file)
#
# This program is free software; you can redistribute it and/or
# modify it under the terms of the GNU General Public License 2
# as published by the Free Software Foundation.
#
# The above copyright notice and this permission notice shall be
# included in all copies or substantial portions of the Software.
#
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

# remove (legacy) plugin-specific polling tasks
rm -f /etc/cron.daily/dwstorecast-poller >/dev/null 2>&1
