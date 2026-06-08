BINARY_NAME=qp
GOBUILD=go build
GOCLEAN=go clean
GOFLAGS=-ldflags="-s -w"
OUTPUT_DIR=build

.PHONY: all build clean install build-all

all: build

build:
	$(GOBUILD) $(GOFLAGS) -o $(BINARY_NAME) .

build-all:
	mkdir -p $(OUTPUT_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(GOFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(GOFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(GOFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-mac-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(GOFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-mac-arm64 .
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(GOFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-windows-amd64.exe .

install: build
	install -m 755 $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

release: clean build-all
	cd $(OUTPUT_DIR) && \
	for f in $(BINARY_NAME)-*; do \
		tar czf "$$f.tar.gz" "$$f" && \
		sha256sum "$$f.tar.gz" > "$$f.tar.gz.sha256"; \
	done
	@echo "Release archives created in $(OUTPUT_DIR)/"

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf $(OUTPUT_DIR)

run:
	$(GOBUILD) -o $(BINARY_NAME) . && ./$(BINARY_NAME)

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy
