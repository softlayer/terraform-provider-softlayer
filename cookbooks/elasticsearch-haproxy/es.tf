provider "softlayer" {}

data "softlayer_vlan" "esk_vlan" {
    number          = "${var.private_vlan_number}"
    router_hostname = "${var.private_router_hostname}"
}

data "softlayer_ssh_key" "esk_key" {
    label = "${var.ssh_key_label}"
}

resource "softlayer_virtual_guest" "esk-node" {
    count                = "${var.node_count}"
    hostname             = "esk-node${count.index+1}"
    domain               = "demo.com"
    os_reference_code    = "UBUNTU_LATEST"
    datacenter           = "${var.datacenter}"
    private_network_only = true
    cores                = 1
    memory               = 1024
    local_disk           = true

    ssh_key_ids = [
        "${data.softlayer_ssh_key.esk_key.id}"
    ]

    private_vlan_id = "${data.softlayer_vlan.esk_vlan.id}"
    private_subnet  = "${data.softlayer_vlan.esk_vlan.subnets.0}"

    provisioner "file" {
        source = "${var.kibana_package}"
        destination = "/tmp/kibana.tar.gz"
    }

    provisioner "remote-exec" {
        script = "es.sh"
    }
}

resource "softlayer_virtual_guest" "haproxy" {
    hostname = "esk-haproxy"
    domain = "demo.com"
    os_reference_code = "UBUNTU_LATEST"
    datacenter = "${var.datacenter}"
    private_network_only = false
    cores = 1
    memory = 1024
    local_disk = true
    user_metadata = "${join(" ", softlayer_virtual_guest.esk-node.*.ipv4_address_private)}"

    ssh_key_ids = [
        "${data.softlayer_ssh_key.esk_key.id}"
    ]

    provisioner "remote-exec" {
        script = "haproxy.sh"
    }
}

output "load_balancer_ip_address" {
    value = "${softlayer_virtual_guest.haproxy.ipv4_address}"
}
