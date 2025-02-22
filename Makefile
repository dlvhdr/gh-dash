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

.PHONY: docs-prepare
docs-prepare:
	cd docs && hugo mod get

.PHONY: docs
docs:
	cd docs && hugo server

.PHONY: mock
mock:
		killgrave --config ./imposters/config.yml &
		FF_MOCK_DATA=true go run . --debug

.PHONY: mock-stop
mock-stop:
	killall killgrave

