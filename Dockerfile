FROM golang:1.10-alpine AS builder

WORKDIR /go/src/github.com/alileza/frog

COPY . ./

RUN apk add --update make git
RUN make build

# ---

FROM alpine

COPY --from=builder /go/src/github.com/alileza/frog /bin/frog

ENTRYPOINT  [ "/bin/frog" ]
CMD         [ "--config.file=/frog.yml" ]
