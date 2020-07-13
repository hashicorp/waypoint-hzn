package main

import (
	"context"
	"crypto/tls"
	"log"
	"net"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/horizon/pkg/control"
	"github.com/hashicorp/horizon/pkg/grpc/lz4"
	hznpb "github.com/hashicorp/horizon/pkg/pb"
	"github.com/jinzhu/gorm"
	"github.com/sethvargo/go-envconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/hashicorp/waypoint-hzn/pkg/server"
)

func main() {
	L := hclog.New(&hclog.LoggerOptions{
		Name:  "wpr",
		Level: hclog.Trace,
	})
	ctx := context.Background()

	L.Info("reading configuration from env vars")
	var cfg config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		log.Fatalf("error reading configuration: %s", err)
		return
	}

	// Apply migrations
	if cfg.MigrationsApply {
		L.Info("applying migrations")

		m, err := migrate.New("file://"+cfg.MigrationsPath, cfg.DatabaseUrl)
		if err != nil {
			log.Fatalf("error creating migrater: %s", err)
		}

		err = m.Up()
		if err != nil {
			if err != migrate.ErrNoChange {
				log.Fatalf("error running migrations: %s", err)
			}
		}

		m.Close()
	}

	L.Info("connecting to database")
	db, err := gorm.Open("postgres", cfg.DatabaseUrl)
	if err != nil {
		log.Fatal(err)
		return
	}

	opts := []grpc.DialOption{
		grpc.WithPerRPCCredentials(control.Token(cfg.ControlToken)),
		grpc.WithDefaultCallOptions(grpc.UseCompressor(lz4.Name)),
	}

	if cfg.ControlInsecure {
		opts = append(opts, grpc.WithInsecure())
	} else {
		creds := credentials.NewTLS(&tls.Config{})
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	L.Info("dialing horizon control plane", "addr", cfg.ControlAddr)
	gcc, err := grpc.Dial(cfg.ControlAddr, opts...)
	if err != nil {
		log.Fatal(err)
		return
	}

	ln, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer ln.Close()

	L.Info("starting server")
	err = server.Run(
		server.WithContext(ctx),
		server.WithLogger(L),
		server.WithGRPC(ln),
		server.WithDB(db),
		server.WithHznControl(hznpb.NewControlManagementClient(gcc)),
		server.WithDomain(cfg.Domain),
	)
	if err != nil {
		log.Fatal(err)
	}
}

type config struct {
	ListenAddr      string `env:"LISTEN_ADDR,default=:24030"`
	ControlAddr     string `env:"CONTROL_ADDR"`
	ControlInsecure bool   `env:"CONTROL_INSECURE"`
	ControlToken    string `env:"CONTROL_TOKEN"`
	DatabaseUrl     string `env:"DATABASE_URL"`
	Domain          string `env:"DOMAIN"`
	MigrationsApply bool   `env:"MIGRATIONS_APPLY"`
	MigrationsPath  string `env:"MIGRATIONS_PATH"`
}
