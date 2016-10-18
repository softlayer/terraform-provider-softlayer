#!/bin/bash

function provision() {
	apt-get install curl -y

	curl -L https://gist.githubusercontent.com/renier/54c854394dd853df9b080cbe834d0911/raw/ca55e84d08b8f7c2491c56264eddf17f5c1a3cfb/harden_ubuntu.sh | bash -s

	ufw allow 9200/tcp

	# Configure HAProxy
	apt-get install haproxy -y
	sed -i -e '/^ENABLED=/d' /etc/default/haproxy
	echo 'ENABLED=1' >> /etc/default/haproxy
	printf "

frontend es_lb
	bind *:9200
	default_backend es_cluster
backend es_cluster
	balance roundrobin
	mode http
" >> /etc/haproxy/haproxy.cfg

	mkdir ~/meta && mount /dev/xvdh1 ~/meta && pushd ~/meta
	sed -e 's/^\["//' meta.js | sed -e 's/"\]$//' > vars.txt
	for i in `cat vars.txt`; do
		printf "	server vm1 ${i}:9200 check\n" >> /etc/haproxy/haproxy.cfg
	done
	popd && umount ~/meta && rmdir ~/meta

	service haproxy reload
}

trap "" HUP
provision > provision.log 2>&1 &