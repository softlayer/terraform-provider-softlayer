#!/bin/bash

apt-get update -y
apt-get install curl -y
apt-get install elasticsearch -y

# Fix https://bugs.launchpad.net/ubuntu/+source/elasticsearch/+bug/1529941
cp /etc/init.d/elasticsearch /tmp/elasticsearch.init.bkup
sed 's/test "\$START_DAEMON/#&/' < /tmp/elasticsearch.init.bkup > /etc/init.d/elasticsearch

# reconfigure elasticsearch to listen on eth1, eth0 and lo
echo "network.host: [_eth0_, _eth1_, _local_]" > /etc/elasticsearch/elasticsearch.yml

systemctl daemon-reload
systemctl restart elasticsearch
systemctl enable elasticsearch.service

# Install Kibana
mkdir -p /opt/kibana
cd /opt/kibana
cat /tmp/kibana.tar.gz | tar -xz --strip-components 1
adduser --shell /bin/false --no-create-home kibana
printf "[Unit]
Description=Kibana

[Service]
Type=simple
User=kibana
Group=kibana
# Load env vars from /etc/default/ and /etc/sysconfig/ if they exist.
# Prefixing the path with '-' makes it try to load, but if the file doesn't
# exist, it continues onward.
EnvironmentFile=-/etc/default/kibana
EnvironmentFile=-/etc/sysconfig/kibana
ExecStart=/opt/kibana/bin/kibana
Restart=always
WorkingDirectory=/

[Install]
WantedBy=multi-user.target" > /etc/systemd/system/kibana.service
service kibana start
systemctl enable kibana.service
