FROM golang:1.15.6-alpine3.12

RUN apk add --no-cache curl

RUN mkdir build
COPY . build/
WORKDIR build

# Build the binary
RUN GO111MODULE=on go mod vendor && GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -mod=vendor -installsuffix cgo -o ./bin/main ./main.go \
    && chmod +x ./bin/main

ENTRYPOINT ["./bin/main"]
