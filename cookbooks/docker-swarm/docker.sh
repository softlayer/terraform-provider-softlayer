#!/bin/sh
apt-get update -y
apt-get install docker.io -y
apt-get install curl -y
curl -L curl -L http://bit.ly/2ejTiG7 | bash -s
ufw allow 2377/tcp
ufw allow 4789/tcp
ufw allow 7946/tcp
public_ip=`ifconfig eth1 | grep "inet addr" | cut -d ':' -f 2 | cut -d ' ' -f 1`
docker swarm init --listen-addr $public_ip:2377 --advertise-addr $public_ip
