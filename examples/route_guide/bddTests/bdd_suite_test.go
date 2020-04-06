package bddTests

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"

	pb "google.golang.org/grpc/examples/route_guide/routeguide"
)

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BDD Test Suite")
}

var (
	clt    pb.RouteGuideClient
	ctx    context.Context
	cancel context.CancelFunc
	conn   *grpc.ClientConn
	err    error
)

var _ = BeforeSuite(func() {
	conn, err = grpc.Dial("localhost:10000", grpc.WithInsecure(), grpc.WithBlock())
	Expect(err).NotTo(HaveOccurred())

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)

	clt = pb.NewRouteGuideClient(conn)
})

var _ = AfterSuite(func() {
	conn.Close()
	cancel()
})
