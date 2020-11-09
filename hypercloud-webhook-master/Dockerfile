FROM golang:1.13-alpine

WORKDIR /usr/src/app
COPY . .
#RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-s' -o main .
ENTRYPOINT ["/usr/src/app/start.sh"]
