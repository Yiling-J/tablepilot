.PHONY: test lint gen run build build-ui install snapshots tauri-dev

test:
	go test ./...
lint:
	golangci-lint run
gen:
	go generate ./...
run:
	go run main.go --config config.toml
build:
	go build -o tablepilot
build-ui:
	cd ui && pnpm build
install:
	go install
snapshots:
	cd tests/cli && go run snapshot.go
tauri-dev:
	go build -o "build/tablepilot-$(shell go run host/host.go)"
	cd ui && pnpm tauri dev
