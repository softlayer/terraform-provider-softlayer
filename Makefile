TEST?=$$(go list ./... | grep -v '/vendor/')
VETARGS?=-all
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)
PLUGIN=provider-softlayer

default: test

tools:
	@go get github.com/kardianos/govendor
	@go get github.com/mitchellh/gox
	@go get golang.org/x/tools/cmd/cover

# bin generates the releaseable binary for your os and architecture
bin: fmtcheck vet tools
	@sh -c "'$(CURDIR)/scripts/build.sh'"

# meant as a pre-step before publishing cross-platform binaries
bins: fmtcheck vet tools
	gox -os="linux darwin windows" -arch="amd64 arm" 

# test runs the unit tests
test: fmtcheck vet
	TF_ACC= go test $(TEST) $(TESTARGS) -timeout=30s -parallel=4

# testacc runs acceptance tests
# e.g make testacc TESTARGS="-run TestAccSoftLayerScaleGroup_Basic"
testacc: fmtcheck vet
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

# testrace runs the race checker
testrace: fmtcheck vet
	TF_ACC= go test -race $(TEST) $(TESTARGS)

cover: tools
	go test $(TEST) -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm coverage.out

# vet runs the Go source code static analysis tool `vet` to find
# any common errors.
vet:
	@echo "go tool vet $(VETARGS) ."
	@go tool vet $(VETARGS) $$(ls -d */ | grep -v vendor) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

.PHONY: bin bins default test vet fmt fmtcheck tools
