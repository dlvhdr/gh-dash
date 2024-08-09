.PHONY: run
run:
		@go run .

.PHONY: run-debug
run-debug:
		go run . --debug

.PHONY: test
test:
		go test -v ./...
