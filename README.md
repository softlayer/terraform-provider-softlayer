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

## Development

### Setup

Make sure you have your [_$GOPATH_ environment variable set](https://golang.org/doc/code.html#GOPATH). $GOPATH should be
set to the terraform-provider-softlayer directory.

_$GOPATH/bin_ should also be in your _$PATH_ (e.g. `export PATH=$GOPATH/bin:PATH`).

To retrieve dependencies run these commands from the root directory of terraform-provider-softlayer.

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
