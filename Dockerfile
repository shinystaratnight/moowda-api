FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

ADD app .
ADD src/moowda/public/index.html public/index.html

EXPOSE 8000

CMD ["./api"]