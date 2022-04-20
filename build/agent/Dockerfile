# syntax=docker/dockerfile:1
FROM golang:1.17 as builder

WORKDIR /build

ENV CGO_ENABLED=0 
ENV GOOS=linux

COPY . .
RUN cd cmd/agent;go build -o /runner -mod mod -a .

FROM openjdk:11.0.14.1-slim

RUN apt-get update && \
    apt-get -y install curl && \
    apt-get purge && \
    curl -Ls https://sh.jbang.dev | bash -s - app setup && \
    export PATH="$HOME/.jbang/bin:$PATH" && \
    jbang app install --name karate com.intuit.karate:karate-core:1.1.0 && \
    karate -h

ENV PATH="$HOME/.jbang/bin:$PATH"
COPY --from=builder /runner /bin/runner
ENTRYPOINT ["/bin/runner"]
