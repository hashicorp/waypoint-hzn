package server

import (
	"context"

	"github.com/hashicorp/horizon/pkg/dbx"
	hznpb "github.com/hashicorp/horizon/pkg/pb"

	"github.com/hashicorp/waypoint-hzn/pkg/models"
	"github.com/hashicorp/waypoint-hzn/pkg/pb"
)

func (s *service) RegisterGuestAccount(
	ctx context.Context,
	req *pb.RegisterGuestAccountRequest,
) (*pb.RegisterGuestAccountResponse, error) {
	accountId := hznpb.NewULID()

	// Register the token with the control server
	ctr, err := s.HznControl.CreateToken(ctx, &hznpb.CreateTokenRequest{
		Account: &hznpb.Account{
			AccountId: accountId,
			Namespace: hznNamespace,
		},
		Capabilities: []hznpb.TokenCapability{
			{
				Capability: hznpb.SERVE,
			},
			{
				Capability: hznpb.CONNECT,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// Register the account in our database
	err = dbx.Check(s.DB.Create(&models.Registration{
		AccountId: accountId.Bytes(),
	}))
	if err != nil {
		return nil, err
	}

	return &pb.RegisterGuestAccountResponse{Token: ctr.Token}, nil
}
