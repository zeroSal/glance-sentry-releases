BINARY_NAME=glance-sentry-releases
BUILD_DIR=build

export CGO_ENABLED=0

.PHONY: test build build-dev build-staging go-mod-tidy go-mod-vendor build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64

test:
	@echo "[*] Running tests..."
	@mkdir -p coverage
	go test ./... -coverpkg=./... -coverprofile=coverage/coverage.cov
	go tool cover -func=coverage/coverage.cov

go-mod-tidy:
	@echo "[*] Running go mod tidy..."
	@go mod tidy

go-mod-vendor: go-mod-tidy
	@echo "[*] Creating vendor directory..."
	@go mod vendor

build-linux-amd64:
	@echo "[*] Building Linux amd64..."
	BUILD_DATE=$$(date -u +%Y-%m-%d); \
	GOOS=linux GOARCH=amd64 GOFLAGS="-mod=vendor" go build -trimpath -ldflags "-s -w -X main.Version=$$VERSION -X main.Channel=$$CHANNEL -X main.BuildDate=$$BUILD_DATE" -o $(BUILD_DIR)/$(BINARY_NAME)-$$VERSION-$$CHANNEL-linux-amd64 main.go

build-linux-arm64:
	@echo "[*] Building Linux arm64..."
	BUILD_DATE=$$(date -u +%Y-%m-%d); \
	GOOS=linux GOARCH=arm64 GOFLAGS="-mod=vendor" go build -trimpath -ldflags "-s -w -X main.Version=$$VERSION -X main.Channel=$$CHANNEL -X main.BuildDate=$$BUILD_DATE" -o $(BUILD_DIR)/$(BINARY_NAME)-$$VERSION-$$CHANNEL-linux-arm64 main.go

build-darwin-amd64:
	@echo "[*] Building Darwin amd64..."
	BUILD_DATE=$$(date -u +%Y-%m-%d); \
	GOOS=darwin GOARCH=amd64 GOFLAGS="-mod=vendor" go build -trimpath -ldflags "-s -w -X main.Version=$$VERSION -X main.Channel=$$CHANNEL -X main.BuildDate=$$BUILD_DATE" -o $(BUILD_DIR)/$(BINARY_NAME)-$$VERSION-$$CHANNEL-darwin-amd64 main.go

build-darwin-arm64:
	@echo "[*] Building Darwin arm64..."
	BUILD_DATE=$$(date -u +%Y-%m-%d); \
	GOOS=darwin GOARCH=arm64 GOFLAGS="-mod=vendor" go build -trimpath -ldflags "-s -w -X main.Version=$$VERSION -X main.Channel=$$CHANNEL -X main.BuildDate=$$BUILD_DATE" -o $(BUILD_DIR)/$(BINARY_NAME)-$$VERSION-$$CHANNEL-darwin-arm64 main.go

build: go-mod-vendor
	ifndef VERSION
		$(error VERSION environment variable not set)
	endif
	ifndef CHANNEL
		$(error CHANNEL environment variable not set)
	endif
		@echo "[*] Building production version $$VERSION-$$CHANNEL..."
		$(MAKE) build-linux-amd64
		$(MAKE) build-linux-arm64
		$(MAKE) build-darwin-amd64
		$(MAKE) build-darwin-arm64
		@echo "[*] All builds done"

build-staging: go-mod-tidy
	@echo "[*] Building staging version..."
	$(MAKE) build-linux-amd64 VERSION=0.0.1 CHANNEL=staging
	$(MAKE) build-linux-arm64 VERSION=0.0.1 CHANNEL=staging
	$(MAKE) build-darwin-amd64 VERSION=0.0.1 CHANNEL=staging
	$(MAKE) build-darwin-arm64 VERSION=0.0.1 CHANNEL=staging

build-dev: go-mod-tidy
	@echo "[*] Building dev version..."
	@VERSION=0.0.1; \
	CHANNEL=dev; \
	BUILD_DATE=$$(date -u +%Y-%m-%d); \
	GOFLAGS="-mod=vendor"; \
	go build -ldflags "-X main.Version=$$VERSION -X main.Channel=$$CHANNEL -X main.BuildDate=$$BUILD_DATE" -o $(BUILD_DIR)/$(BINARY_NAME)-dev main.go
