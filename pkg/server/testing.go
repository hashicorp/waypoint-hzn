package server

import (
	"context"
	"net"

	"github.com/mitchellh/go-testing-interface"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/hashicorp/waypoint-hzn/pkg/pb"
)

// TestServer starts a server and returns various data such as the client
// for that server. We use t.Cleanup to ensure resources are automatically
// cleaned up.
func TestServer(t testing.T) *TestServerData {
	require := require.New(t)

	// Listen on a random port
	ln, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(err)
	t.Cleanup(func() { ln.Close() })

	// Create the server
	ctx, cancel := context.WithCancel(context.Background())
	go Run(
		WithContext(ctx),
		WithGRPC(ln),
	)
	t.Cleanup(func() { cancel() })

	// Connect, this should retry in the case Run is not going yet
	conn, err := grpc.DialContext(ctx, ln.Addr().String(),
		grpc.WithBlock(),
		grpc.WithInsecure(),
	)
	require.NoError(err)
	t.Cleanup(func() { conn.Close() })

	return &TestServerData{
		Addr:   ln.Addr().String(),
		Client: pb.NewWaypointHznClient(conn),
	}
}

type TestServerData struct {
	Addr   string
	Client pb.WaypointHznClient
}
