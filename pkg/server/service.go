package server

import (
	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/waypoint-hzn/pkg/pb"
)

// service implements pb.WaypointHznServer.
type service struct {
	L hclog.Logger
}

var _ pb.WaypointHznServer = (*service)(nil)
