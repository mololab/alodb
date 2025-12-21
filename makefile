.PHONY: run build tidy test clean

# Run the application
run:
	go run cmd/main.go

# Build the application
build:
	go build -o bin/alodb cmd/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
