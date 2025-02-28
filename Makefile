.PHONY: test lint gen run build

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
