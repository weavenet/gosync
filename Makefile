base_dir = `pwd`
gopath = "$(base_dir)/vendor:$(GOPATH)"

all: fmt deps test
	@echo "Building."
	@mkdir -p bin/
	@env GOPATH=$(gopath) go build -v -o bin/gosync .
clean:
	@echo "Cleaning."
	@rm -rf bin pkg vendor/pkg
deps:
	@echo "Getting Dependencies."
	@env GOPATH=$(gopath) go get -d -v ./...
fmt:
	@echo "Formatting."
	gofmt -w .
test: deps
	@echo "Testing."
	@env GOPATH=$(gopath) go test ./gosync/...

.PHONY: all clean deps fmt test
