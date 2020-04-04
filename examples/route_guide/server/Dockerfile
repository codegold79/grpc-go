FROM photon:3.0

# Install go, mostly per instructions at https://golang.org/doc/install
# GOPATH was set because root is being used, not a user.
RUN tdnf -y install git wget tar gzip && \
    wget https://dl.google.com/go/go1.14.1.linux-amd64.tar.gz && \
    tar -C /usr/local/ -xzf go1.14.1.linux-amd64.tar.gz
ENV PATH=${PATH}:/usr/local/go/bin:/go/bin GOPATH=/go

# Install gRPC and route_guide example
WORKDIR /go/src/github.com/grpc/
RUN go get google.golang.org/grpc && \
    git clone https://github.com/grpc/grpc-go.git

# Copy over tests into GOPATH
WORKDIR /go/src/github.com/grpc/grpc-go/examples/route_guide
COPY . .

# Build the server, start it, and run the tests.
WORKDIR /go/src/github.com/grpc/grpc-go/examples/route_guide/server
RUN go build server.go
ENTRYPOINT [ "./server"]
