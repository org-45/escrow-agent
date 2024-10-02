#!/bin/bash

if [ -z "$1" ]; then
  echo "Usage: $0 <password>"
  exit 1
fi

PASSWORD=$1

sleep 5

gnome-terminal -- bash -c "echo $PASSWORD | sudo -S docker logs -f escrow-agent-app; exec bash"
gnome-terminal -- bash -c "echo $PASSWORD | sudo -S docker logs -f escrow-agent-db; exec bash"
