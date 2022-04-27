FROM golang:1.18 AS builder

RUN apt-get update && apt-get install git && rm -rf /var/lib/apt/lists/*

WORKDIR /go/src/pod-server

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 go build -o /go/bin/pod-server -ldflags="-s -w" /go/src/pod-server/*.go

FROM scratch

COPY --from="builder" /go/bin/pod-server /pod-server

ENTRYPOINT ["/pod-server"]
