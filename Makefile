all: fmt deps test
	@echo "Building."
	@mkdir -p bin/
	go build -v -o bin/gosync .
deps:
	@echo "Getting Dependencies."
	go get -d -v ./...
test: deps
	@echo "Testing."
	go test ./...
fmt:
	@echo "Formatting."
	gofmt -w -tabs=false -tabwidth=2 .
