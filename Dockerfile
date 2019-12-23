FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git ca-certificates upx binutils
RUN adduser -D -g '' appuser

COPY . $GOPATH/src/github.com/gutmensch/goweck/
WORKDIR $GOPATH/src/github.com/gutmensch/goweck/

RUN go get -u github.com/go-bindata/go-bindata/... \
  && go-bindata -o bindata.go asset/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOFLAGS=-mod=vendor go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/goweck \
  && strip -s /go/bin/goweck \
  && upx /go/bin/goweck

FROM scratch

ARG LISTEN_PORT="8081"

COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/bin/goweck /go/bin/goweck

ENV ZONEINFO=/usr/local/go/lib/time/zoneinfo.zip
ENV LISTEN=":${LISTEN_PORT}"

USER appuser

EXPOSE ${LISTEN_PORT}

ENTRYPOINT ["/go/bin/goweck"]
