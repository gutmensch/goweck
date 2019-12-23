FROM golang:alpine as builder

RUN apk update && apk add --no-cache git ca-certificates upx binutils

RUN adduser -D -g '' appuser

COPY . $GOPATH/src/github.com/gutmensch/goweck/
WORKDIR $GOPATH/src/github.com/gutmensch/goweck/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOFLAGS=-mod=vendor go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/goweck \
  && strip -s /go/bin/goweck \
  && upx /go/bin/goweck

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/bin/goweck /go/bin/goweck

USER appuser

ENTRYPOINT ["/go/bin/goweck"]
