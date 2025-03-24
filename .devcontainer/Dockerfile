FROM golang:1.23-bookworm
ENV GOLANG_VERSION 1.23.7

RUN apt update && apt install -y \
    build-essential \
    git \
    curl \
    ssh 

# Install Viam RDK CLI
RUN ARCH=$(go env GOARCH) && \
    curl -o /usr/local/bin/viam https://storage.googleapis.com/packages.viam.com/apps/viam-cli/viam-cli-stable-linux-${ARCH} && \
    chmod a+rx /usr/local/bin/viam

ENV PATH=/usr/local/go/bin:${PATH}
RUN groupadd -g 1000 user
RUN useradd -u 1000 -g 1000 -m -s /bin/bash user
RUN usermod -aG sudo user
RUN echo "user ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers
RUN cat /etc/passwd

WORKDIR /workspace
USER ${USERNAME}
CMD ["/bin/bash"]
