# terraform-provider-softlayer

## Install

```
$ go get github.com/softlayer/terraform-provider-softlayer
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

You should have the correct source in your _$GOPATH_ for both terraform and softlayer-go. _$GOPATH/bin_ should also be in your _$PATH_.

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
