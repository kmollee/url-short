FROM golang as builder

ADD . /go/src/github.com/kmollee/url-short/

WORKDIR /go/src/github.com/kmollee/url-short/

RUN go get

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o url-short .

FROM scratch

ENV PORT 8080

COPY --from=builder /go/src/github.com/kmollee/url-short/url-short /app/

WORKDIR /app

CMD ["./url-short"]
