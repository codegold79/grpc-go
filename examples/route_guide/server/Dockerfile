FROM golang:1.14 AS builder

# Install gRPC and repo containing route_guide example
WORKDIR /go/src/github.com/grpc/
RUN go get google.golang.org/grpc && \
    git clone https://github.com/grpc/grpc-go.git

# Build and run the server
WORKDIR /go/src/github.com/grpc/grpc-go/examples/route_guide/server
RUN go build server.go

FROM photon:3.0
WORKDIR /go/src/github.com/grpc/grpc-go/examples/route_guide/server
COPY --from=builder /go/src/github.com/grpc/grpc-go/examples/route_guide/server .

ENTRYPOINT [ "./server"]
