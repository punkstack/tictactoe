# Build your project
build:
	go build -o tictactoe

# Build and run your project for debugging
debug:
	go run main.go

# Run unit tests
test:
	go test ./...

# Clean up generated files (e.g., compiled binaries)
clean:
	rm -f tictactoe

# Define additional targets and actions as needed
# ...

# Define phony targets (targets that don't represent files)
.PHONY: build debug test clean