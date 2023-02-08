MODULE_NAME = eth-parser

TEST_FILES = $(shell find -L * -name '*_test.go' -not -path "vendor/*" -not -name "tools.go")
TEST_PACKAGES = $(dir $(addprefix $(MODULE_NAME)/,$(TEST_FILES)))


all: parser

clean:
	rm -rf bin/

parser:
	go build -mod vendor -o bin/parser

bin/golangci-lint:
	@echo "getting golangci-lint for $$(uname -m)/$$(uname -s)"
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.28.3

lint: bin/golangci-lint
	bin/golangci-lint run -v -c golangci.yml

test:
	go test -cover -mod vendor $(TEST_PACKAGES)

.PHONY: parser
