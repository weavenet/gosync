all: deps test
	@mkdir -p bin/
	go build -v -o bin/gosync .
deps:
	go get -d -v ./...
test: deps
	go test ./...
