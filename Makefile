.DEFAULT_GOAL = build

build:
	go build -o gh-dash gh-dash.go

clean:
	go clean

install:
	gh extension install .
