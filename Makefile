BINARY     := service-control-mcp
CMD        := ./cmd/$(BINARY)
BIN_DIR    := bin
COVERAGE   := coverage.txt

GO         := go
GOLINT     := golangci-lint
MIN_COVER  := 80

.PHONY: all build test lint fmt fmt-check vet ci clean coverage

all: build

build:
	$(GO) build -o $(BIN_DIR)/$(BINARY) $(CMD)

test:
	$(GO) test -count=1 -race -coverprofile=$(COVERAGE) ./...

coverage: test
	@total=$$($(GO) tool cover -func=$(COVERAGE) | grep total | awk '{print $$3}' | tr -d '%'); \
	echo "Total coverage: $${total}%"; \
	if awk "BEGIN { exit !($${total} < $(MIN_COVER)) }"; then \
		echo "Coverage $${total}% is below minimum $(MIN_COVER)%"; \
		exit 1; \
	fi

lint:
	$(GOLINT) run ./...

fmt:
	$(GO) fmt ./...

fmt-check:
	@out=$$(gofmt -l .); \
	if [ -n "$$out" ]; then \
		echo "Files need formatting:"; \
		echo "$$out"; \
		exit 1; \
	fi

vet:
	$(GO) vet ./...

ci: fmt-check vet lint test coverage

clean:
	rm -rf $(BIN_DIR) $(COVERAGE)
