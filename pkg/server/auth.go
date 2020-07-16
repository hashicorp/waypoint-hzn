package server

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/horizon/pkg/token"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *service) checkAuth(ctx context.Context) (*token.ValidToken, error) {
	L := hclog.FromContext(ctx)

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.PermissionDenied,
			"authentication information not presented")
	}

	auth := md["authorization"]

	if len(auth) < 1 {
		return nil, status.Errorf(codes.PermissionDenied,
			"authentication information not presented")
	}

	token, err := token.CheckTokenED25519(auth[0], s.tokenPub)
	if err != nil {
		L.Warn("error checking token signature", "error", err)
		return nil, status.Errorf(codes.PermissionDenied,
			"authentication information not presented")
	}

	account := token.Account()
	if account.Namespace != s.Namespace {
		return nil, status.Errorf(codes.PermissionDenied,
			"invalid token")
	}

	return token, nil
}
