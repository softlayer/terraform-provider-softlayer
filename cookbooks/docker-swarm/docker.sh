#!/bin/sh
apt-get update -y > /dev/null
apt-get install docker.io -y
apt-get install curl -y

ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw --force enable

# Harden SSH
sed -i -e '/^PermitRootLogin/d' /etc/ssh/sshd_config
sed -i -e '/^ChallengeResponseAuthentication/d' /etc/ssh/sshd_config
sed -i -e '/^PasswordAuthentication/d' /etc/ssh/sshd_config
sed -i -e '/^UsePAM/d' /etc/ssh/sshd_config
echo 'PermitRootLogin without-password' >> /etc/ssh/sshd_config
echo 'DebianBanner no' >> /etc/ssh/sshd_config
echo 'ChallengeResponseAuthentication no' >> /etc/ssh/sshd_config
echo 'PasswordAuthentication no' >> /etc/ssh/sshd_config
echo 'UsePAM no' >> /etc/ssh/sshd_config
service ssh restart

# Harden against TCP attacks
printf "
# Ignore ICMP broadcast requests
net.ipv4.icmp_echo_ignore_broadcasts = 1

# Disable source packet routing
net.ipv4.conf.all.accept_source_route = 0
net.ipv6.conf.all.accept_source_route = 0
net.ipv4.conf.default.accept_source_route = 0
net.ipv6.conf.default.accept_source_route = 0

# Ignore send redirects
net.ipv4.conf.all.send_redirects = 0
net.ipv4.conf.default.send_redirects = 0

# Block SYN attacks
net.ipv4.tcp_max_syn_backlog = 2048
net.ipv4.tcp_synack_retries = 2
net.ipv4.tcp_syn_retries = 5

# Log Martians
net.ipv4.conf.all.log_martians = 1
net.ipv4.icmp_ignore_bogus_error_responses = 1

# Ignore ICMP redirects
net.ipv4.conf.all.accept_redirects = 0
net.ipv6.conf.all.accept_redirects = 0
net.ipv4.conf.default.accept_redirects = 0
net.ipv6.conf.default.accept_redirects = 0

# Ignore Directed pings
net.ipv4.icmp_echo_ignore_all = 1
" >> /etc/sysctl.d/10-network-security.conf
service procps start

ufw allow 2377/tcp
ufw allow 4789/tcp
ufw allow 7946/tcp
public_ip=`ifconfig eth1 | grep "inet addr" | cut -d ':' -f 2 | cut -d ' ' -f 1`
docker swarm init --listen-addr $public_ip:2377 --advertise-addr $public_ip
