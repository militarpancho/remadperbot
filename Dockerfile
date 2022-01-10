FROM golang:1.17-alpine

WORKDIR /go/src/remadperbot
COPY . .
RUN go mod tidy && \
	go build -o /go/src/remadperbot/remadperbot


FROM scratch

COPY --from=0 /go/src/remadperbot/remadperbot /go/bin/remadperbot
CMD ["/go/bin/remadperbot"]