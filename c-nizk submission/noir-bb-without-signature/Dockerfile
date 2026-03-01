FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    curl \
    git \
    ca-certificates \
    build-essential \
    pkg-config \
    libssl-dev \
    unzip \
    bash \
    wget \
    jq \
    xz-utils \
    nodejs \
    npm \
    && rm -rf /var/lib/apt/lists/*

ENV GO_VERSION=1.22.5

RUN wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"

RUN curl -L https://raw.githubusercontent.com/noir-lang/noirup/main/install | bash

ENV PATH="/root/.nargo/bin:${PATH}"

RUN noirup

RUN curl -L https://raw.githubusercontent.com/AztecProtocol/aztec-packages/refs/heads/next/barretenberg/bbup/install | bash

ENV PATH="/root/.bb:${PATH}"

RUN bbup

RUN noirup --version
RUN nargo --version
RUN bb --version

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o server .

ENV PORT=4000

EXPOSE 4000

CMD ["./server"]