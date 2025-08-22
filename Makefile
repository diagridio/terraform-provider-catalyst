TEST?=$$(go list ./... |grep -v 'vendor')

default: build

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

doc:
	@cd tools && go generate .

# Please keep targets in alphabetical order
.PHONY: \
	build \
	ci-fmt-check \
	fmt \
	install \
	test \
	testacc \
	doc


