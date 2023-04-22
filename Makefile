.PHONY: check lint-check-deps

check: lint-check-deps
	@echo "[golangci-lint] linting sources"
	@golangci-lint run

lint-check-deps:
	@if [ -z `which golangci-lint` ]; then \
		echo "[go get] installing golangci-lint";\
		GO111MODULE=on go get -u github.com/golangci/golangci-lint/cmd/golangci-lint;\
	fi