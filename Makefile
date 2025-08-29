TEST?=$$(go list ./... |grep -v 'vendor')

MOCKGEN_VERSION ?= v0.4.0

default: build

.PHONY: mock-tools
mock-tools: ## Install mockgen tool
	@GO111MODULE=on go install go.uber.org/mock/mockgen@$(MOCKGEN_VERSION)

generate: mock-tools
	go generate ./...

build: ## Build provider
	@go build .

fmt:
	gofmt -w $(GOFMT_FILES)

install: build ## Install provider
	@go install

testacc: # Run acceptance tests
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc: ## Run acceptance tests
	TF_ACC=1 go test -i $(TEST) -timeout 5m || exit 1

doc:
	@go generate ./...

# Please keep targets in alphabetical order
.PHONY: \
	build \
	ci-fmt-check \
	fmt \
	install \
	test \
	testacc \
	doc


