provider "softlayer" {}

output "cluster_address" {
    value = "http://${softlayer_lb_local.es_lb_vip.ip_address}:${var.port}/"
}

data "softlayer_vlan" "es_vlan" {
    number          = "${var.backend_vlan_number}"
    router_hostname = "${var.backend_primary_router_hostname}"
}

data "softlayer_ssh_key" "es_key" {
    label = "${var.ssh_key_label}"
}

resource "softlayer_lb_local" "es_lb_vip" {
    connections = 250
    datacenter  = "${var.datacenter}"
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
    datacenter           = "${var.datacenter}"
    hourly_billing       = true
    cpu                  = 1
    ram                  = 1024
    disks                = [25]
    local_disk           = true

    ssh_keys = [
        "${data.softlayer_ssh_key.es_key.id}"
    ]

    private_vlan_id = "${data.softlayer_vlan.es_vlan.id}"

    private_subnet = "${var.backend_subnet}"

    provisioner "remote-exec" {
        script = "es.sh"
    }
}

# Add guest as Load Balancer Service (member)
resource "softlayer_lb_local_service" "es_member" {
    count             = "${var.node_count}"
    service_group_id  = "${softlayer_lb_local_service_group.es_lb_sg.service_group_id}"
    ip_address_id     = "${element(softlayer_virtual_guest.es-vm.*.ip_address_id, count.index)}"
    port              = "${var.port}"
    enabled           = true
    health_check_type = "HTTP"
    weight            = 1
}
