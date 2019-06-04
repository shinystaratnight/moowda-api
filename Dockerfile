FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

ADD api .
COPY config/development.yaml ./config/

EXPOSE 8080

CMD ["./api"]