variable node_count {
    default = 2
}

variable datacenter {
    default = "dal06"
}

# Network settings
variable private_vlan_number {
    default = 1170
}

variable private_router_hostname {
    default = "bcr02a.dal06"
}

# SSH Keys
variable ssh_key_label {
    default = "Renier Personal"
}

# Kibana package
variable kibana_package {
    default = "kibana-4.1.11-linux-x64.tar.gz"
}
