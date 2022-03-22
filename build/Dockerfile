FROM golang:alpine as builder

ENV GO111MODULE on
RUN apk update && apk add git && apk add ca-certificates
RUN adduser -D -g '' unprivileged
COPY . $GOPATH/src/github.com/joa/website/
WORKDIR $GOPATH/src/github.com/joa/website/

RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/app

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/bin/app /go/bin/app
USER unprivileged
EXPOSE 8080
ENTRYPOINT ["/go/bin/app"]
