GOPATH := $(shell go env GOPATH)
GOARCH := $(shell go env GOARCH)
GOOS := $(shell go env GOOS)

# The installer is pinned alongside the version it installs: fetching it from
# master means every build runs whatever that branch holds today.
GOLANGCI_VERSION := 2.13.1

all: test

getdeps:
	@mkdir -p ${GOPATH}/bin
	@if ! ${GOPATH}/bin/golangci-lint --version 2>/dev/null | grep -qF " $(GOLANGCI_VERSION) "; then \
		echo "Installing golangci-lint v$(GOLANGCI_VERSION)"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/v$(GOLANGCI_VERSION)/install.sh | sh -s -- -b $(GOPATH)/bin v$(GOLANGCI_VERSION); \
	fi

lint: getdeps
	@echo "Running $@ check"
	@${GOPATH}/bin/golangci-lint cache clean
	@${GOPATH}/bin/golangci-lint run --build-tags kqueue --timeout=10m --config ./.golangci.yml

lint-fix: getdeps
	@echo "Running $@ check"
	@${GOPATH}/bin/golangci-lint cache clean
	@${GOPATH}/bin/golangci-lint run --build-tags kqueue --timeout=10m --config ./.golangci.yml --fix

test: lint
	@echo "Running unit tests"
	@go test -race -tags kqueue ./...

test-ldap: lint
	@echo "Running unit tests for LDAP with LDAP server at '"${LDAP_TEST_SERVER}"'"
	@go test -v -race ./ldap

clean:
	@echo "Cleaning up all the generated files"
	@find . -name '*.test' | xargs rm -fv
	@find . -name '*~' | xargs rm -fv
