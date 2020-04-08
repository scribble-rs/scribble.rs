.DEFAULT: help
.SILENT:
SHELL=bash

help: ## Display usage
	printf "\033[96mScribble.rs\033[0m\n\n"
	grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

are-requirements-ok:
	which go >/dev/null 2>&1 || { echo >&2 "'go' is required.\nPlease install it."; exit 1; }

build: are-requirements-ok ## Build binary file
	go run github.com/markbates/pkger/cmd/pkger -include /resources -include /templates
	go build -o scribblers .
	printf "\033[32mBuild done!\033[0m\n"
