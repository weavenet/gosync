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
	@echo "==> Removing .git, .bzr, and .hg from third_party."
	@find ./vendor -type d -name .git | xargs rm -rf
	@find ./vendor -type d -name .bzr | xargs rm -rf
	@find ./vendor -type d -name .hg | xargs rm -rf
fmt:
	@echo "Formatting."
	gofmt -w .
test: deps
	@echo "Testing."
	@env GOPATH=$(gopath) go test ./gosync/...

.PHONY: all clean deps fmt test
