# terraform-provider-softlayer

[![Chat on gitter](https://img.shields.io/gitter/room/softlayer/terraform-provider-softlayer.svg?maxAge=2592000)](https://gitter.im/softlayer/terraform-provider-softlayer) [![Build Status](https://travis-ci.org/softlayer/terraform-provider-softlayer.svg?branch=master)](https://travis-ci.org/softlayer/terraform-provider-softlayer)

**Deprecated**: Please refer to https://github.com/IBM-Cloud/terraform-provider-ibm

## Install

[Download the binary](https://github.com/softlayer/terraform-provider-softlayer/releases) of the provider.

Create or edit this file to specify the location of the terraform softlayer provider binary:

```
# ~/.terraformrc
providers {
    softlayer = "/path/to/bin/terraform-provider-softlayer"
}
```

## Documentation

Go to the [documentation directory](docs/).

## Development

### Setup

1. Ensure you have a [_$GOPATH_ environment variable set](https://golang.org/doc/code.html#GOPATH).
1. Ensure you have _$GOPATH/bin_ in your _$PATH_ (e.g. `export PATH=$GOPATH/bin:PATH`).
1. Install _terraform-provider-softlayer_.
```
$ go get -u github.com/softlayer/terraform-provider-softlayer
```
1. Get the main dependency:
```
$ go get github.com/hashicorp/terraform
```

The project will exist at `$GOPATH/src/github.com/softlayer/terraform-provider-softlayer`.

### Build

```
make bin
```

### Test

```
make
```

To run the acceptance tests (**warning**: Requires a SoftLayer
account and resources will be provisioned):

```
make testacc
```

### Updating dependencies

We are using [govendor](https://github.com/kardianos/govendor) to manage dependencies just like Terraform. Please see its documentation for additional help.
