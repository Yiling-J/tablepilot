.PHONY: test lint gen run build build-ui install

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
