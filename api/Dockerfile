FROM golang:alpine as builder

RUN mkdir /builder

ADD . /build/

WORKDIR /build

RUN go build -o main .

FROM alpine:latest

RUN adduser -S -D -H -h /app appuser

#  -S → system user (no login shell)
# -D → no password
# -H → no home directory created
# -h /app says → "home directory is /app"

USER appuser

COPY . /app

COPY --from=builder /build/main /app/

WORKDIR /app


EXPOSE 3000

CMD ["./main"]

