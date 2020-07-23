.DEFAULT_GOAL := build

.PHONY: build run run-client

build:
	go build

run: build
	./context_examples

run-client:
	curl -b user=jon localhost:3000/?data=hello
