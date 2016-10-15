#!/bin/bash

apt-get update -y
apt-get install elasticsearch -y

# Fix https://bugs.launchpad.net/ubuntu/+source/elasticsearch/+bug/1529941
cp /etc/init.d/elasticsearch /tmp/elasticsearch.init.bkup
sed 's/test "\$START_DAEMON/#&/' < /tmp/elasticsearch.init.bkup > /etc/init.d/elasticsearch

# reconfigure elasticsearch to listen on eth1, eth0 and lo
cp /tmp/elasticsearch.yml /etc/elasticsearch

systemctl daemon-reload
systemctl restart elasticsearch
