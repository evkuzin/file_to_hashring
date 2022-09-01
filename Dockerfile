FROM golang:1.18.2-stretch

WORKDIR /src

COPY ./go.mod ./go.sum ./file_to_hashring/

WORKDIR /src/file_to_hashring

RUN go mod download

COPY . ./

WORKDIR /src/file_to_hashring

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN make build

FROM alpine:3.16.2

WORKDIR /usr/local/bin

COPY --from=0 /src/file_to_hashring/bin/main ./file_to_hashring

ENTRYPOINT ["/usr/local/bin/file_to_hashring"]