GOWORK := $(abspath examples/workspace/go.work)
BUILDDIR := build
GOIMPORTS_REVISER := go run github.com/incu6us/goimports-reviser/v3@latest
EXAMPLES := $(filter-out examples/workspace,$(wildcard examples/*/main.go))
EXAMPLES := $(patsubst examples/%/main.go,%,$(EXAMPLES))

.PHONY: build clean test vet format format-check check

build: clean
	@mkdir -p $(BUILDDIR)
	@$(foreach ex,$(EXAMPLES),\
		echo "Building examples/$(ex) -> $(BUILDDIR)/$(ex)" && \
		GOWORK=$(GOWORK) go build -o $(BUILDDIR)/$(ex) ./examples/$(ex) &&) true

clean:
	rm -rf $(BUILDDIR)

test:
	go test -v ./...

vet:
	go vet ./...
	@$(foreach ex,$(EXAMPLES),\
		echo "Vetting examples/$(ex)" && \
		GOWORK=$(GOWORK) go vet ./examples/$(ex) &&) true

format-check:
	@echo "Checking formatting..."
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)
	@echo "Checking imports..."
	@$(GOIMPORTS_REVISER) -list-diff -set-exit-status ./...

format:
	@echo "Fixing formatting..."
	@gofmt -w .
	@echo "Fixing imports..."
	@$(GOIMPORTS_REVISER) ./...

check: format-check vet test
