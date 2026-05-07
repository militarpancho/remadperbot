FROM golang:1.26-alpine3.23 AS builder

ARG OS=linux
ARG ARCH=amd64

RUN apk add --no-cache ca-certificates git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=${OS} GOARCH=${ARCH} go build -trimpath -tags netgo -ldflags="-s -w" -o /out/remadperbot .

FROM alpine:3.23

RUN apk add --no-cache ca-certificates \
	&& adduser -D -H -u 10001 -s /sbin/nologin appuser

COPY --from=builder /out/remadperbot /usr/local/bin/remadperbot

USER appuser:appuser
ENTRYPOINT ["/usr/local/bin/remadperbot"]
