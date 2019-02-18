ARG GO_VERSION=1.11

FROM golang:${GO_VERSION}-stretch

ARG VAULT_VERSION=1.0.2
ENV CGO_ENABLED=0
ENV GOARCH=amd64
ENV GO111MODULE=on
ENV BUILD_DIR=/BUILD

# Install utility programs
RUN \
	apt-get update -y && \
	apt-get install -y \
		wget \
		unzip \
		jq && \
	apt-get clean

# Install Vault CLI
RUN \
	mkdir /tmp/build && \
	cd /tmp/build && \
	wget "https://releases.hashicorp.com/vault/${VAULT_VERSION}/vault_${VAULT_VERSION}_linux_${GOARCH}.zip" && \
	unzip -d /bin vault_${VAULT_VERSION}_linux_${GOARCH}.zip && \
	cd /tmp && \
	rm -rf /tmp/build

# Build plugin
#
# Using the new Go module system means we can
# build outside of the $GOPATH.

COPY . /SRC
WORKDIR /SRC
RUN \
	go build \
	-mod=vendor \
	-o ${BUILD_DIR}/vault-plugin-auth-cloudfoundry

# This script should be mounted at runtime
# into a $PATH dir
ENTRYPOINT ["entrypoint.bash"]
