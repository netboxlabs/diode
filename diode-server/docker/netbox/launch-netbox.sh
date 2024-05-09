#!/bin/bash

UNIT_CONFIG="${UNIT_CONFIG-/opt/netbox/nginx-unit.json}"
# Also used in "nginx-unit.json"
UNIT_SOCKET="/opt/unit/unit.sock"

load_configuration() {
  MAX_WAIT=10
  WAIT_COUNT=0
  while [ ! -S $UNIT_SOCKET ]; do
    if [ $WAIT_COUNT -ge $MAX_WAIT ]; then
      echo "⚠️ No control socket found; configuration will not be loaded."
      return 1
    fi

    WAIT_COUNT=$((WAIT_COUNT + 1))
    echo "⏳ Waiting for control socket to be created... (${WAIT_COUNT}/${MAX_WAIT})"

    sleep 1
  done

  # even when the control socket exists, it does not mean unit has finished initialisation
  # this curl call will get a reply once unit is fully launched
  curl --silent --output /dev/null --request GET --unix-socket $UNIT_SOCKET http://localhost/

  echo "⚙️ Applying configuration from $UNIT_CONFIG"

  RESP_CODE=$(
    curl \
      --silent \
      --output /dev/null \
      --write-out '%{http_code}' \
      --request PUT \
      --data-binary "@${UNIT_CONFIG}" \
      --unix-socket $UNIT_SOCKET \
      http://localhost/config
  )
  if [ "$RESP_CODE" != "200" ]; then
    echo "⚠️ Could no load Unit configuration"
    kill "$(cat /opt/unit/unit.pid)"
    return 1
  fi

  echo "✅ Unit configuration loaded successfully"
}

reload_netbox() {
  if [ "$RELOAD_NETBOX_ON_DIODE_PLUGIN_CHANGE" == "true" ]; then
    netbox_diode_plugin_md5=$(find /opt/netbox/netbox/netbox_diode_plugin -type f -name "*.py" -exec md5sum {} + | md5sum | awk '{print $1}')

    while true; do
      new_md5=$(find /opt/netbox/netbox/netbox_diode_plugin -type f -name "*.py" -exec md5sum {} + | md5sum | awk '{print $1}')
      if [ "$netbox_diode_plugin_md5" != "$new_md5" ]; then
        echo "🔄 Reloading NetBox"
        curl --silent --output /dev/null --unix-socket /opt/unit/unit.sock -X GET http://localhost/control/applications/netbox/restart
        netbox_diode_plugin_md5=$new_md5
      fi
      sleep 1
    done
  fi
}

load_configuration &

reload_netbox &

exec unitd \
  --no-daemon \
  --control unix:$UNIT_SOCKET \
  --pid /opt/unit/unit.pid \
  --log /dev/stdout \
  --statedir /opt/unit/state/ \
  --tmpdir /opt/unit/tmp/ \
  --user unit \
  --group root
