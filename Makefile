TEST?=$$(go list ./... | grep -v '/vendor/')
VETARGS?=-all
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)
PLUGIN=provider-softlayer

default: test vet

tools:
	go get -u github.com/kardianos/govendor

# bin generates the releaseable binaries for Terraform
bin: fmtcheck generate
	@gox -help 2>/dev/null; if [ $$? -ne 0 ]; then \
		go get -u github.com/mitchellh/gox; \
	fi
	@sh -c "'$(CURDIR)/scripts/build.sh'"

# test runs the unit tests
test: fmtcheck generate
	TF_ACC= go test $(TEST) $(TESTARGS) -timeout=30s -parallel=4

# testacc runs acceptance tests
testacc: fmtcheck generate
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

# testrace runs the race checker
testrace: fmtcheck generate
	TF_ACC= go test -race $(TEST) $(TESTARGS)

cover:
	@go tool cover 2>/dev/null; if [ $$? -eq 3 ]; then \
		go get -u golang.org/x/tools/cmd/cover; \
	fi
	go test $(TEST) -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm coverage.out

# vet runs the Go source code static analysis tool `vet` to find
# any common errors.
vet:
	@go tool vet 2>/dev/null ; if [ $$? -eq 3 ]; then \
		go get golang.org/x/tools/cmd/vet; \
	fi
	@echo "go tool vet $(VETARGS) ."
	@go tool vet $(VETARGS) $$(ls -d */ | grep -v vendor) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

generate:
	go generate $$(go list ./... | grep -v /vendor/)

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

.PHONY: bin default generate test vet fmt fmtcheck tools
