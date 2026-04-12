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

git:
	git add .
	git commit -m "logic improvements"
	git push

g:
	make git

prod:
	make g || true
	git pull
	git checkout prod
	git reset --hard main
	git push origin prod -f
	git checkout main