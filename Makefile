EXAMPLES := $(shell go list ./... | grep "examples")

.PHONY: lint examples

lint-pkgs:
	GO111MODULE=off go get -u honnef.co/go/tools/cmd/staticcheck
	GO111MODULE=off go get -u github.com/client9/misspell/cmd/misspell

lint:
	$(exit $(go fmt ./... | wc -l))
	go vet ./...
	find . -type f -name "*.go" | xargs misspell -error -locale US
	staticcheck $(go list ./... | grep -v ent/privacy)

examples:
	$(foreach var, $(EXAMPLES), go build -buildmode=plugin -o ./.bin/${var}.so $(var);)
