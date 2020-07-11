package server

import (
	"context"

	"github.com/hashicorp/waypoint-hzn/pkg/pb"
)

func (s *service) RegisterHostname(
	ctx context.Context,
	req *pb.RegisterHostnameRequest,
) (*pb.RegisterHostnameResponse, error) {
	return nil, nil
}
