provider "softlayer" {}

resource "softlayer_ssh_key" "ssh_key" {
  label = "${var.ssh_label}"
  public_key = "${file(var.ssh_key_path)}"
}

resource "softlayer_virtual_guest" "gateway" {
  hostname = "gateway-test"
  domain = "danube.cf"
  cores = 1
  memory = 2048
  datacenter = "${var.datacenter}"
  os_reference_code = "UBUNTU_16_64"
  local_disk = true
  hourly_billing = true

  ssh_key_ids = ["${softlayer_ssh_key.ssh_key.id}"]

  provisioner "remote-exec" {
    script = "gateway.sh"
  }
}

resource "softlayer_virtual_guest" "node" {
  hostname = "gw-node"
  domain = "danube.cf"
  cores = 1
  memory = 2048
  datacenter = "${var.datacenter}"
  os_reference_code = "UBUNTU_16_64"
  private_network_only = true
  hourly_billing = true
  private_vlan_id = "${softlayer_virtual_guest.gateway.private_vlan_id}"
  private_subnet = "${softlayer_virtual_guest.gateway.private_subnet}"

  ssh_key_ids = ["${softlayer_ssh_key.ssh_key.id}"]

  provisioner "remote-exec" {
    inline = [
      "echo '${softlayer_virtual_guest.gateway.ipv4_address_private}' > /tmp/proxy_ip"
    ]
  }

  provisioner "remote-exec" {
    script = "node.sh"
  }
}
