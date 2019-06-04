FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

ADD api .
ADD config/development.yaml .

EXPOSE 8000

CMD ["./api"]