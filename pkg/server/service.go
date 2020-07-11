package server

import (
	hznpb "github.com/hashicorp/horizon/pkg/pb"
	"github.com/jinzhu/gorm"

	"github.com/hashicorp/waypoint-hzn/pkg/pb"
)

// service implements pb.WaypointHznServer.
type service struct {
	DB         *gorm.DB
	Domain     string
	HznControl hznpb.ControlManagementClient
}

var _ pb.WaypointHznServer = (*service)(nil)
