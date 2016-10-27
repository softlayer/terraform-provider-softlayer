# terraform-provider-softlayer

[![Chat on gitter](https://img.shields.io/gitter/room/softlayer/terraform-provider-softlayer.svg?maxAge=2592000)](https://gitter.im/softlayer/terraform-provider-softlayer) [![Build Status](https://travis-ci.org/softlayer/terraform-provider-softlayer.svg?branch=master)](https://travis-ci.org/softlayer/terraform-provider-softlayer)

## Install

```
$ go get -u github.com/softlayer/terraform-provider-softlayer
```

Alternatively, you can now also [download binaries](https://github.com/softlayer/terraform-provider-softlayer/releases) of the provider.

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
1. Install _terraform-provider-softlayer_ as described above in the **install** section.
1. Get the main dependencies:
```
$ go get github.com/softlayer/softlayer-go/...
$ go get github.com/hashicorp/terraform
```

The project will exist at _$GOPATH/src/github.com/softlayer/terraform-provider-softlayer_.

### Build

```
make bin
```

### Test

```
make
```

### Updating dependencies

We are using [govendor](https://github.com/kardianos/govendor) to manage dependencies just like Terraform. Please see its documentation for additional help.
