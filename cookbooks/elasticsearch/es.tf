provider "softlayer" {}

output "cluster_address" {
    value = "http://${softlayer_lb_local.es_lb_vip.ip_address}:${var.port}/"
}

data "softlayer_vlan" "es_vlan" {
    number =          "${var.backend_vlan_number}"
    router_hostname = "${var.backend_primary_router_hostname}"
}

resource "softlayer_ssh_key" "es_key" {
    label      = "ES Demo Key"
    public_key = "${file("~/.ssh/es_id_rsa.pub")}"
}

resource "softlayer_lb_local" "es_lb_vip" {
    connections = 250
    datacenter  = "wdc01"
    dedicated   = false
}

resource "softlayer_lb_local_service_group" "es_lb_sg" {
    load_balancer_id = "${softlayer_lb_local.es_lb_vip.id}"
    allocation       = 100
    port             = "${var.port}"
    routing_method   = "ROUND_ROBIN"
    routing_type     = "HTTP"
}

resource "softlayer_virtual_guest" "es-vm" {
    count                = "${var.node_count}"
    name                 = "es-vm${count.index+1}"
    domain               = "demo.com"
    os_reference_code    = "UBUNTU_LATEST"
    datacenter           = "wdc01"
    network_speed        = 100
    hourly_billing       = true
    private_network_only = false
    cpu                  = 1
    ram                  = 1024
    disks                = [25]
    local_disk           = true

    ssh_keys = [
        "${softlayer_ssh_key.es_key.id}"
    ]

    private_vlan_id = "${data.softlayer_vlan.es_vlan.id}"

    back_end_subnet = "${var.backend_subnet}"

    # Note: the private key cannot be password protected
    connection {
        host        = "${self.ipv4_address}"
        user        = "root"
        private_key = "${file("~/.ssh/es_id_rsa")}"
    }

    provisioner "file" {
        source      = "elasticsearch.yml"
        destination = "/tmp/elasticsearch.yml"
    }

    provisioner "file" {
        source      = "es.sh"
        destination = "/tmp/es.sh"
    }

    provisioner "remote-exec" {
        inline = [
            "sudo chmod 700 /tmp/es.sh",
            "sudo /tmp/es.sh"
        ]
    }
}

# Add guest as Load Balancer Service (member)
resource "softlayer_lb_local_service" "es_member" {
    count             = "${var.node_count}"
    service_group_id  = "${softlayer_lb_local_service_group.es_lb_sg.service_group_id}"
    ip_address_id     = "${element(softlayer_virtual_guest.es-vm.*.ip_address_id, count.index)}"
    port              = 9200
    enabled           = true
    health_check_type = "HTTP"
    weight            = 1
}
