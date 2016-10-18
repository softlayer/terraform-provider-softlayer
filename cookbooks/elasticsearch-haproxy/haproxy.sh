#!/bin/bash

function provision() {
	apt-get install curl -y

	curl -L https://gist.githubusercontent.com/renier/54c854394dd853df9b080cbe834d0911/raw/ca55e84d08b8f7c2491c56264eddf17f5c1a3cfb/harden_ubuntu.sh | bash -s

	ufw allow 9200/tcp
	ufw allow 5601/tcp

	# Configure HAProxy
	apt-get install haproxy -y
	sed -i -e '/^ENABLED=/d' /etc/default/haproxy
	echo 'ENABLED=1' >> /etc/default/haproxy

	mkdir ~/meta && mount /dev/xvdh1 ~/meta && pushd ~/meta
	sed -e 's/^\["//' meta.js | sed -e 's/"\]$//' > vars.txt

	printf "

frontend es_lb
	bind *:9200
	default_backend es_cluster
backend es_cluster
	balance roundrobin
	mode http
" >> /etc/haproxy/haproxy.cfg

	for i in `cat vars.txt`; do
		printf "	server vm1 ${i}:9200 check\n" >> /etc/haproxy/haproxy.cfg
	done

	printf "

frontend kibana_lb
	bind *:5601
	default_backend kibana_cluster
backend kibana_cluster
	balance roundrobin
	mode http
" >> /etc/haproxy/haproxy.cfg
	for i in `cat vars.txt`; do
		printf "	server k1 ${i}:5601 check\n" >> /etc/haproxy/haproxy.cfg
	done

	popd && umount ~/meta && rmdir ~/meta

	service haproxy reload
	systemctl enable haproxy.service
}

provision > provisioning.log 2>&1