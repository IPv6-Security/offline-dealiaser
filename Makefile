ifeq ($(OS),Windows_NT)
  EXECUTABLE_EXTENSION := .exe
else
  EXECUTABLE_EXTENSION :=
endif

GO_FILES = $(shell find . -type f -name '*.go')
TEST_MODULES ?= 

all: build

# Test currently only runs on the modules folder because some of the
# third-party libraries in lib (e.g. http) are failing.
test:
	cd lib/output/test && go test -v ./...
	cd modules && go test -v ./...

update: clean
	go clean -cache
	go get -u all
	go clean -cache
	cd cmd/aliasv6 && go build -a && cd ../..
	rm -f aliasv6
	ln -s cmd/aliasv6/aliasv6$(EXECUTABLE_EXTENSION) aliasv6
	go mod tidy

gofmt:
	goimports -w -l $(GO_FILES)

build: $(GO_FILES)
	cd cmd/aliasv6 && go build && cd ../..
	rm -f aliasv6
	ln -s cmd/aliasv6/aliasv6$(EXECUTABLE_EXTENSION) aliasv6

clean:
	cd cmd/aliasv6 && go clean
	rm -f aliasv6
