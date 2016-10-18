provider "softlayer" {}

data "softlayer_vlan" "es_vlan" {
    number          = "${var.private_vlan_number}"
    router_hostname = "${var.private_router_hostname}"
}

data "softlayer_ssh_key" "es_key" {
    label = "${var.ssh_key_label}"
}

resource "softlayer_virtual_guest" "es-vm" {
    count                = "${var.node_count}"
    name                 = "es-vm${count.index+1}"
    domain               = "demo.com"
    os_reference_code    = "UBUNTU_LATEST"
    datacenter           = "${var.datacenter}"
    private_network_only = true
    cpu                  = 1
    ram                  = 1024
    local_disk           = true

    ssh_keys = [
        "${data.softlayer_ssh_key.es_key.id}"
    ]

    private_vlan_id = "${data.softlayer_vlan.es_vlan.id}"
    private_subnet  = "${data.softlayer_vlan.es_vlan.subnets.0}"

    # Note: the private key cannot be password protected
    connection {
        host  = "${self.ipv4_address_private}"
    }

    provisioner "remote-exec" {
        script = "es.sh"
    }
}

resource "softlayer_virtual_guest" "haproxy" {
    name = "haproxy"
    domain = "demo.com"
    os_reference_code = "UBUNTU_LATEST"
    datacenter = "${var.datacenter}"
    private_network_only = false
    cpu = 1
    ram = 1024
    local_disk = true
    user_data = "${join(" ", softlayer_virtual_guest.es-vm.*.ipv4_address_private)}"

    ssh_keys = [
        "${data.softlayer_ssh_key.es_key.id}"
    ]

    connection {
        host = "${self.ipv4_address}"
    }

    provisioner "remote-exec" {
        script = "haproxy.sh"
    }
}

output "load_balancer_ip_address" {
    value = "${softlayer_virtual_guest.haproxy.ipv4_address}"
}
