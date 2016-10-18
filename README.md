# terraform-provider-softlayer

[![Chat on gitter](https://img.shields.io/gitter/room/softlayer/terraform-provider-softlayer.svg?maxAge=2592000)](https://gitter.im/softlayer/terraform-provider-softlayer) [![Build Status](https://travis-ci.org/softlayer/terraform-provider-softlayer.svg?branch=master)](https://travis-ci.org/softlayer/terraform-provider-softlayer)

## Install

```
$ go get -u github.com/softlayer/terraform-provider-softlayer
```

Create or edit this file to specify the location of the terraform softlayer provider binary:

```
# ~/.terraformrc
providers {
    softlayer = "/path/to/bin/terraform-provider-softlayer"
}
```

## Documentation

Go to the [documentation directory](docs/).

#### `softlayer_global_ip`

Provides Global Ip's containing all the information needed to add a global ip. This allows Global Ip's to be created, updated and deleted.
For additional details please refer to [API documentation](http://sldn.softlayer.com/reference/services/SoftLayer_Network_Subnet_IpAddress_Global).

##### Example Usage

```hcl
resource "softlayer_global_ip" "test_global_ip " {
    routes_to = "119.81.82.163"
}
```

##### Argument Reference

The following arguments are supported:

* `routes_to` | *string* - (Required) Create a new transaction to modify a global IP route.

Field `routes_to`is editable.

##### Attributes Reference

The following attributes are exported:

* `id` - id of the new global ip
* `ip_address` - ip address of the new global ip

## Development

### Setup

Make sure you have your [_$GOPATH_ environment variable set](https://golang.org/doc/code.html#GOPATH).

_$GOPATH/bin_ should also be in your _$PATH_ (e.g. `export PATH=$GOPATH/bin:PATH`).

Now you need to get the main dependencies.

To get _softlayer-go_:

```
go get github.com/softlayer/softlayer-go/...
```

To get _terraform_:

```
go get github.com/hashicorp/terraform
```

### Build

```
make bin
```

### Test

```
make test
```

### Updating dependencies

We are using [govendor](https://github.com/kardianos/govendor) to manage dependencies just like Terraform. Please see its documentation for additional help.
