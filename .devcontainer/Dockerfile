FROM golang:1.23-alpine

RUN apk add --no-cache bash git make curl openssh

# Install Viam RDK CLI
RUN ARCH=$(go env GOARCH) && \
    curl -o /usr/local/bin/viam https://storage.googleapis.com/packages.viam.com/apps/viam-cli/viam-cli-stable-linux-${ARCH} && \
    chmod a+rx /usr/local/bin/viam

ENV PATH=/usr/local/go/bin:${PATH}
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=1000
RUN addgroup -g ${USER_GID} ${USERNAME} \
    && adduser -u ${USER_UID} -G ${USERNAME} -s /bin/bash -D ${USERNAME}

WORKDIR /workspace
USER ${USERNAME}
CMD ["/bin/bash"]
