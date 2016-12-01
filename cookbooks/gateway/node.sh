#!/bin/bash

# Change default gateway here
gateway=`cat /tmp/proxy_ip`
ip route replace default via $gateway
sed -i -e "s/gateway \\(10\\..*\\)/gateway $gateway/g" /etc/network/interfaces

apt-get install curl -y

# Hardening
apt-get remove --purge snapd -y
apt-get autoremove -y
apt-get clean
curl -L http://bit.ly/2fYvdF0 | bash -s -- guarionx
