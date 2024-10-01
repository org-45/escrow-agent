#!/bin/bash

sudo docker-compose up -d

sleep 5


gnome-terminal -- bash -c "sudo docker logs -f escrow-agent-app; exec bash"

gnome-terminal -- bash -c "sudo docker logs -f escrow-agent-db; exec bash"

gnome-terminal -- bash -c "sudo docker logs -f escrow-agent-frontend; exec bash"

gnome-terminal -- bash -c "sudo docker logs -f minio; exec bash"

gnome-terminal -- bash -c "sudo docker logs -f mc; exec bash"

