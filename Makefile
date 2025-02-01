.PHONY: run
run:
		@go run .

.PHONY: run-debug
run-debug:
		go run . --debug

.PHONY: test
test:
		go test -v ./...

.PHONY: check-nerd-font
check-nerd-font:
	nerdfix check --quiet $$(fd --extension go)

.PHONY: fix-nerd-font
fix-nerd-font:
	nerdfix fix --quiet --format=json $$(fd --extension go)
