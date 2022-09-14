ENV_FILE :=
SHELL := /bin/bash
LICENSES_DIR := licenses

export

.PHONY: license
license:
	go mod tidy
	rm -rf "${LICENSES_DIR}"
	mkdir -p "${LICENSES_DIR}"
	go-licenses save . --force --save_path "${LICENSES_DIR}" --alsologtostderr
	chmod +w -R "${LICENSES_DIR}"

.PHONY: goreleaser
goreleaser:
	goreleaser release --snapshot --rm-dist

.PHONY: test
test:
	go test -v ./...
