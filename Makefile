GOWORK := $(abspath examples/workspace/go.work)
BUILDDIR := build
GOIMPORTS_REVISER := go run github.com/incu6us/goimports-reviser/v3@latest
EXAMPLES := $(filter-out examples/workspace,$(wildcard examples/*/main.go))
EXAMPLES := $(patsubst examples/%/main.go,%,$(EXAMPLES))

GLAP_SOURCES := $(wildcard *.go)
EXAMPLE_BINARIES := $(addprefix $(BUILDDIR)/,$(EXAMPLES))

.PHONY: build clean test test-examples vet format format-check check

build: $(EXAMPLE_BINARIES)

$(BUILDDIR)/%: examples/%/main.go $(GLAP_SOURCES)
	@mkdir -p $(BUILDDIR)
	@echo "Building examples/$* -> $@"
	@GOWORK=$(GOWORK) go build -o $@ ./examples/$*

clean:
	rm -rf $(BUILDDIR)

test:
	go test -v ./...

test-examples: build
	@failures=0; \
	for ex in $(EXAMPLES); do \
		if [ -x examples/$$ex/test.sh ]; then \
			echo "Testing examples/$$ex"; \
			examples/$$ex/test.sh $(abspath $(BUILDDIR))/$$ex || failures=$$((failures + 1)); \
		fi; \
	done; \
	if [ $$failures -ne 0 ]; then \
		echo "$$failures example test(s) failed"; \
		exit 1; \
	fi

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

check: format-check vet test test-examples
