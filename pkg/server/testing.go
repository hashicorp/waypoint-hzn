package server

import (
	"context"
	"net"

	hzncontrol "github.com/hashicorp/horizon/pkg/control"
	"github.com/hashicorp/horizon/pkg/grpc/lz4"
	hznpb "github.com/hashicorp/horizon/pkg/pb"
	hzntest "github.com/hashicorp/horizon/pkg/testutils/central"
	"github.com/mitchellh/go-testing-interface"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/hashicorp/waypoint-hzn/internal/testsql"
	"github.com/hashicorp/waypoint-hzn/pkg/pb"
)

// TestServer starts a server and returns various data such as the client
// for that server. We use t.Cleanup to ensure resources are automatically
// cleaned up.
func TestServer(t testing.T, opts ...Option) *TestServerData {
	require := require.New(t)

	// Create the server
	var data TestServerData
	data.readyCh = make(chan struct{})
	go Run(
		append(
			append([]Option{}, opts...),
			testWithDefaults(t, &data),
		)...,
	)

	// Wait for it to start
	<-data.readyCh

	// Connect, this should retry in the case Run is not going yet
	conn, err := grpc.DialContext(context.Background(), data.Addr,
		grpc.WithBlock(),
		grpc.WithInsecure(),
	)
	require.NoError(err)
	t.Cleanup(func() { conn.Close() })
	data.Client = pb.NewWaypointHznClient(conn)

	return &data
}

type TestServerData struct {
	Addr   string
	Client pb.WaypointHznClient
	Hzn    *hzntest.DevSetup

	readyCh chan struct{}
}

// TestGuestAccount registers a guest account and returns a context that
// can be used with auth information for future API calls.
func TestGuestAccount(t testing.T, client pb.WaypointHznClient) grpc.CallOption {
	resp, err := client.RegisterGuestAccount(
		context.Background(), &pb.RegisterGuestAccountRequest{
			ServerId: "A",
		},
	)
	require.NoError(t, err)
	require.NotEmpty(t, resp.Token)

	return grpc.PerRPCCredentials(hzncontrol.Token(resp.Token))
}

func testWithDefaults(t testing.T, data *TestServerData) Option {
	return func(opts *options) {
		defer close(data.readyCh)
		testWithContext(t, opts)
		testWithListener(t, opts, data)
		testWithDB(t, opts, data)
		testWithHzn(t, opts, data)

		if opts.Domain == "" {
			opts.Domain = "waypoint-hzn.localhost"
		}
	}
}

func testWithContext(t testing.T, opts *options) {
	// Setup the context
	if opts.Context == nil {
		opts.Context = context.Background()
	}

	// We need the context to be cancellable
	ctx, cancel := context.WithCancel(opts.Context)
	opts.Context = ctx
	t.Cleanup(func() { cancel() })
}

func testWithListener(t testing.T, opts *options, data *TestServerData) {
	if opts.GRPCListener == nil {
		// Listen on a random port
		ln, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		t.Cleanup(func() { ln.Close() })
		opts.GRPCListener = ln
	}

	data.Addr = opts.GRPCListener.Addr().String()
}

func testWithDB(t testing.T, opts *options, data *TestServerData) {
	if opts.DB == nil {
		opts.DB = testsql.TestPostgresDB(t, "waypoint_hzn_test")
	}
}

func testWithHzn(t testing.T, opts *options, data *TestServerData) {
	if opts.HznControl == nil {
		// Create the test server. On test end we close the channel which quits
		// the Horizon test server.
		setupCh := make(chan *hzntest.DevSetup, 1)
		closeCh := make(chan struct{})
		t.Cleanup(func() { close(closeCh) })
		go hzntest.Dev(t, func(setup *hzntest.DevSetup) {
			setupCh <- setup
			<-closeCh
		})
		data.Hzn = <-setupCh

		// We need a management token for our namespace
		token, err := data.Hzn.ControlServer.GetManagementToken(context.Background(), hznNamespace)
		require.NoError(t, err)

		// New connection that uses this token
		conn, err := grpc.Dial(data.Hzn.ServerAddr,
			grpc.WithInsecure(),
			grpc.WithPerRPCCredentials(hzncontrol.Token(token)),
			grpc.WithDefaultCallOptions(grpc.UseCompressor(lz4.Name)),
		)
		require.NoError(t, err)
		t.Cleanup(func() { conn.Close() })
		opts.HznControl = hznpb.NewControlManagementClient(conn)
	}
}
