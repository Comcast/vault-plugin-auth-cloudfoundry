GOLANG_VERSION := latest
SRC_PATH := $(shell go list -e)
TARGET := $(shell basename ${SRC_PATH})
TARGET_DIR := dev/vault/plugins
PWD_DIR := $(shell pwd)
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

print-%  : ; @echo $* = $($*)

build: $(TARGET_DIR)/$(TARGET)
	@true

.PHONY: clean
clean:
	rm -f "${TARGET_DIR}/${TARGET}"

$(TARGET_DIR)/$(TARGET): $(SRC)
	docker run \
		--env CGO_ENABLED=0 \
		--env GOOS=linux \
		--env GOARCH=amd64 \
		--env GO111MODULE=on \
		--rm \
		--volume="$(shell pwd):/go/src/${SRC_PATH}:ro" \
		--volume="$(shell pwd)/${TARGET_DIR}:/OUTPUT" \
		--workdir="/go/src/${SRC_PATH}" \
		"golang:${GOLANG_VERSION}" \
			go build \
				-mod=vendor \
				-o="/OUTPUT/${TARGET}"
