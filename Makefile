.DEFAULT: help
.SILENT:
SHELL=bash

help: ## Display usage
	printf "\033[96mScribble.rs\033[0m\n\n"
	grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

is-go-installed:
	which go >/dev/null 2>&1 || { echo >&2 "'go' is required.\nPlease install it."; exit 1; }

build: is-go-installed ## Build binary file
	CGO_ENABLED=0 go build -ldflags="-w -s" -o scribblers .
	printf "\033[32mBuild done!\033[0m\n"
