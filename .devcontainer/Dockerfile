# syntax=docker/dockerfile:1.6
FROM golang:1.26-bookworm

# ---- Base dev utilities ----
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
      ca-certificates \
      curl \
      git \
      bash \
      openssh-client \
      unzip \
      nano \
    && rm -rf /var/lib/apt/lists/*

# Install Viam RDK CLI
RUN ARCH=$(go env GOARCH) && \
    curl -o /usr/local/bin/viam https://storage.googleapis.com/packages.viam.com/apps/viam-cli/viam-cli-stable-linux-${ARCH} && \
    chmod a+rx /usr/local/bin/viam

# ---- Create non-root vscode user ----
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=1000

RUN groupadd --gid ${USER_GID} ${USERNAME} \
  && useradd --uid ${USER_UID} --gid ${USER_GID} -m ${USERNAME} \
  && mkdir -p /workspaces /home/vscode/go /go/pkg/mod \
  && chown -R ${USERNAME}:${USERNAME} /workspaces /home/vscode/go /go

ENV GOPATH="/home/vscode/go"
ENV GOMODCACHE="/go/pkg/mod"
WORKDIR /workspaces
USER ${USERNAME}

# ---- Install common Go tools ----
RUN go install golang.org/x/tools/gopls@latest \
 && go install honnef.co/go/tools/cmd/staticcheck@latest \
 && go install golang.org/x/tools/cmd/goimports@latest

CMD ["/bin/bash"]