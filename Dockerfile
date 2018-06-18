FROM golang:alpine as builder

RUN apk add --no-cache git && apk add -U --no-cache ca-certificates
ADD . /go/src/github.com/pajk/go-http-proxy/
RUN CGO_ENABLED=0 GOOS=linux go build -a --ldflags '-extldflags "-static"' github.com/pajk/go-http-proxy

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/go-http-proxy /bin/http_proxy

CMD ["/bin/http_proxy"]
