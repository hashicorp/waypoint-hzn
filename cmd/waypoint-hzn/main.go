package main

import (
	"context"
	"crypto/tls"
	"flag"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"net/url"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/horizon/pkg/grpc/lz4"
	grpctoken "github.com/hashicorp/horizon/pkg/grpc/token"
	hznpb "github.com/hashicorp/horizon/pkg/pb"
	"github.com/hashicorp/waypoint-hzn/pkg/server"
	"github.com/jinzhu/gorm"
	"github.com/sethvargo/go-envconfig"
	"google.golang.org/grpc"
)

var (
	flagDev = flag.Bool("dev", false, "dev mode to run locally")
)

func main() {
	flag.Parse()

	L := hclog.New(&hclog.LoggerOptions{
		Name:  "waypoint-hzn",
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
	if cfg.MigrationsApply || *flagDev {
		if !filepath.IsAbs(cfg.MigrationsPath) {
			path, err := filepath.Abs(cfg.MigrationsPath)
			if err != nil {
				log.Fatalf("error determining migration path: %s", err)
			}

			cfg.MigrationsPath = path
		}

		L.Info("applying migrations", "path", cfg.MigrationsPath)
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
	u, err := url.Parse(cfg.DatabaseUrl)
	if err != nil {
		log.Fatal(err)
		return
	}
	db, err := gorm.Open("postgres", migrate.FilterCustomQuery(u).String())
	if err != nil {
		log.Fatal(err)
		return
	}

	opts := []grpc.DialOption{
		grpc.WithPerRPCCredentials(grpctoken.Token(cfg.ControlToken)),
		grpc.WithDefaultCallOptions(grpc.UseCompressor(lz4.Name)),
	}

	if cfg.ControlInsecure {
		opts = append(opts, grpc.WithInsecure())
	} else {
		var creds credentials.TransportCredentials
		if cfg.ControlInsecureSkipVerify {
			creds = credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
		} else {
			creds = credentials.NewTLS(&tls.Config{})
		}
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
		server.WithGRPC(ln, cfg.ReflectionEnabled),
		server.WithDB(db),
		server.WithHznControl(hznpb.NewControlManagementClient(gcc)),
		server.WithDomain(cfg.Domain),
	)
	if err != nil {
		log.Fatal(err)
	}
}

type config struct {
	ListenAddr                string `env:"LISTEN_ADDR,default=:24030"`
	ControlAddr               string `env:"CONTROL_ADDR,default=127.0.0.1:24401"`
	ControlInsecure           bool   `env:"CONTROL_INSECURE,default=1"`
	ControlToken              string `env:"CONTROL_TOKEN,default=aabbcc"`
	DatabaseUrl               string `env:"DATABASE_URL"`
	Domain                    string `env:"DOMAIN,default=waypoint.localdomain"`
	MigrationsApply           bool   `env:"MIGRATIONS_APPLY"`
	MigrationsPath            string `env:"MIGRATIONS_PATH,default=./migrations"`
	ControlInsecureSkipVerify bool   `env:"CONTROL_INSECURE_SKIP_VERIFY,default=0"`
	ReflectionEnabled         bool   `env:"CONTROL_GRPC_REFLECTION_API,default=0"`
}
