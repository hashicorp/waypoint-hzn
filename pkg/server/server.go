package server

import (
	"context"
	"net"

	"github.com/hashicorp/go-hclog"
	hznpb "github.com/hashicorp/horizon/pkg/pb"
	"github.com/jinzhu/gorm"
	"github.com/oklog/run"
)

// hznNamespace is the namespace we use for all our Horizon API calls.
const hznNamespace = "/waypoint"

// Run initializes and starts the server. This will block until the server
// exits (by cancelling the associated context set with WithContext or due
// to an unrecoverable error).
func Run(opts ...Option) error {
	var cfg options
	for _, opt := range opts {
		opt(&cfg)
	}

	// Set defaults
	if cfg.Context == nil {
		cfg.Context = context.Background()
	}
	if cfg.Logger == nil {
		cfg.Logger = hclog.L()
	}
	if cfg.Namespace == "" {
		cfg.Namespace = hznNamespace
	}

	// Setup our run group since we're going to be starting multiple
	// goroutines for all the servers that we want to live/die as a group.
	var group run.Group

	// We first add an actor that just returns when the context ends. This
	// will trigger the rest of the group to end since a group will not exit
	// until any of its actors exit.
	ctx, cancelCtx := context.WithCancel(cfg.Context)
	group.Add(func() error {
		<-ctx.Done()
		return ctx.Err()
	}, func(error) { cancelCtx() })

	// Setup our gRPC server.
	if err := grpcInit(&group, &cfg); err != nil {
		return err
	}

	// Run!
	return group.Run()
}

// Option configures Run
type Option func(*options)

// options configure a server and are set by users only using the exported
// Option functions.
type options struct {
	// Context is the context to use for the server. When this is cancelled,
	// the server will be gracefully shutdown.
	Context context.Context

	// Logger is the logger to use. This will default to hclog.L() if not set.
	Logger hclog.Logger

	// GRPCListener will setup the gRPC server. If this is nil, then a
	// random loopback port will be chosen. The gRPC server must run since it
	// serves the HTTP endpoints as well.
	GRPCListener net.Listener
	// GRPCReflection flag determines whether GRPC should expose a reflection endpoint
	// so that clients can see the endpoints.
	GRPCReflection bool

	// PostgreSQL DB connection.
	DB *gorm.DB

	// Client to Horizon control client
	HznControl hznpb.ControlManagementClient

	// Domain to use
	Domain string

	// Horizon namespace for all accounts
	Namespace      string
}

// WithContext sets the context for the server. When this context is cancelled,
// the server will be shut down.
func WithContext(ctx context.Context) Option {
	return func(opts *options) { opts.Context = ctx }
}

// WithLogger sets the logger.
func WithLogger(log hclog.Logger) Option {
	return func(opts *options) { opts.Logger = log }
}

// WithGRPC sets the GRPC listener. This listener must be closed manually
// by the caller. Prior to closing the listener, it is recommended that you
// cancel the context set with WithContext and wait for Run to return.
func WithGRPC(ln net.Listener, reflectionEnabled bool) Option {

	return func(opts *options) {
		opts.GRPCListener = ln
		opts.GRPCReflection = reflectionEnabled
	}
}

// WithDB sets the DB connection.
func WithDB(db *gorm.DB) Option {
	return func(opts *options) { opts.DB = db }
}

// WithHznControl
func WithHznControl(client hznpb.ControlManagementClient) Option {
	return func(opts *options) { opts.HznControl = client }
}

// WithDomain
func WithDomain(d string) Option {
	return func(opts *options) { opts.Domain = d }
}

// WithNamespace
func WithNamespace(ns string) Option {
	return func(opts *options) { opts.Namespace = ns }
}
