# syntax=docker/dockerfile:1
FROM golang:1.18

WORKDIR /build

ENV CGO_ENABLED=0 
ENV GOOS=linux

COPY . .
RUN cd cmd/agent;go build -o /runner -mod mod -a .

FROM adoptopenjdk/openjdk11:jdk-11.0.18_10-slim
ARG KARATEVERSION=1.3.1

RUN apt-get update && \
    apt-get -y install curl && \
    apt-get purge && \
    useradd -m -d /home/karate -s /bin/bash -u 1001 -r -g root karate

USER 1001
WORKDIR /home/karate

RUN curl -Ls "https://github.com/karatelabs/karate/releases/download/v$KARATEVERSION/karate-$KARATEVERSION.jar" --output karate.jar && \
    java -jar karate.jar -h

COPY --from=0 /runner /bin/runner
ENTRYPOINT ["/bin/runner"]
