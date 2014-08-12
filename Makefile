all: fmt deps test
	@echo "Building."
	@mkdir -p bin/
	go build -v -o bin/gosync .
deps:
	@echo "Getting Dependencies."
	go get -d -v ./...
test: deps
	@echo "Testing."
	go test -cover ./...
fmt:
	@echo "Formatting."
	gofmt -w .
