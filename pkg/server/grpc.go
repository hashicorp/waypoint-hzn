package server

import (
	"google.golang.org/grpc/reflection"
	"time"

	hznpb "github.com/hashicorp/horizon/pkg/pb"
	petname "github.com/hashicorp/waypoint-hzn/internal/pkg/golang-petname"
	"github.com/oklog/run"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/hashicorp/waypoint-hzn/pkg/pb"
)

// grpcInit initializes the gRPC server and adds it to the run group.
func grpcInit(group *run.Group, opts *options) error {
	log := opts.Logger.Named("grpc")

	var so []grpc.ServerOption

	/*
		if opts.AuthChecker != nil {
			so = append(so,
				grpc.ChainUnaryInterceptor(authUnaryInterceptor(opts.AuthChecker)),
				grpc.ChainStreamInterceptor(authStreamInterceptor(opts.AuthChecker)),
			)
		}
	*/

	so = append(so,
		grpc.ChainUnaryInterceptor(
			// Insert our logger and also log req/resp
			logUnaryInterceptor(log, false),
		),
		grpc.ChainStreamInterceptor(
			// Insert our logger and log
			logStreamInterceptor(log, false),
		),
	)

	s := grpc.NewServer(so...)

	// Get our public key
	tokenInfo, err := opts.HznControl.GetTokenPublicKey(opts.Context, &hznpb.Noop{})
	if err != nil {
		return err
	}

	// Setup petname randomization
	petname.NonDeterministicMode()

	// Register our server
	pb.RegisterWaypointHznServer(s, &service{
		DB:         opts.DB,
		Domain:     opts.Domain,
		Namespace:  opts.Namespace,
		HznControl: opts.HznControl,
		tokenPub:   tokenInfo.PublicKey,
		Logger:     opts.Logger,
	})

	// Register our health check
	hs := grpchealth.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	healthpb.RegisterHealthServer(s, hs)

	// Add our gRPC server to the run group
	group.Add(func() error {
		// Set our status to healthy
		hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

		// Serve traffic
		ln := opts.GRPCListener
		log.Info("starting gRPC server", "addr", ln.Addr().String())
		if opts.GRPCReflection {
			reflection.Register(s)
		}
		return s.Serve(ln)
	}, func(err error) {
		// Graceful in a goroutine so we can timeout
		gracefulCh := make(chan struct{})
		go func() {
			defer close(gracefulCh)
			log.Info("shutting down gRPC server")
			s.GracefulStop()
		}()

		select {
		case <-gracefulCh:

		// After a timeout we just forcibly exit. Our gRPC endpoints should
		// be fairly quick and their operations are atomic so we just kill
		// the connections after a few seconds.
		case <-time.After(2 * time.Second):
			s.Stop()
		}
	})

	return nil
}
