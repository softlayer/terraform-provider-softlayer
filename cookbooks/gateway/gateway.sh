#!/bin/bash

apt-get install curl -y

# Hardening
apt-get remove --purge snapd -y
apt-get autoremove -y
apt-get clean
curl -L http://bit.ly/2fYvdF0 | bash -s -- guarionx

# Become gateway for private vlan (eth0)
ufw default allow routed
echo 1 > /proc/sys/net/ipv4/ip_forward
echo 'net.ipv4.ip_forward=1' >> /etc/sysctl.d/99-sysctl.conf
echo 1 > /proc/sys/net/ipv6/conf/default/forwarding
echo 1 > /proc/sys/net/ipv6/conf/all/forwarding
echo 'net.ipv6.conf.default.forwarding=1' >> /etc/sysctl.d/99-sysctl.conf
echo 'net.ipv6.conf.all.forwarding=1' >> /etc/sysctl.d/99-sysctl.conf
sed -i -e '/DEFAULT_FORWARD_POLICY=/d' /etc/default/ufw
echo 'DEFAULT_FORWARD_POLICY="ACCEPT"' >> /etc/default/ufw
echo 'net/ipv4/ip_forward=1' >> /etc/ufw/sysctl.conf
echo 'net/ipv6/conf/default/forwarding=1' >> /etc/ufw/sysctl.conf
echo 'net/ipv6/conf/all/forwarding=1' >> /etc/ufw/sysctl.conf
printf "
*nat
:POSTROUTING ACCEPT [0:0]
-A POSTROUTING -s 10.0.0.0/8 -o eth1 -j MASQUERADE
COMMIT

$(cat /etc/ufw/before.rules)
" > /etc/ufw/before.rules
ufw disable && ufw --force enable
